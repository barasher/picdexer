package setup

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"mime/multipart"
	"net/http"
	"text/template"
	"time"
)

const (
	picdexerIndex = "picdexer"
	syncOnDateIndex="sync-on-date"
)

//go:embed assets/kibana.ndjson
var kibanaComponentsPayload string

type ESManagerInterface interface {
	MappingAlreadyExist(client *http.Client, index string) (bool, error)
	DeleteMapping(client *http.Client, index string) error
	PutMapping(client *http.Client, index string, mapping string) error
}

type Setup struct {
	esUrl  string
	kibUrl string
	fsUrl  string
}

func logReader(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error while reading response body: %w", err)
	}
	log.Debug().Msgf("Response body: %s", string(b))
	return nil
}

func NewSetup(esUrl string, kibUrl string, fsUrl string) (*Setup, error) {
	return &Setup{esUrl: esUrl, kibUrl: kibUrl, fsUrl: fsUrl}, nil
}

func (s *Setup) setupElasticsearch(m ESManagerInterface) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	if err := s.setupIndex(&client, m, picdexerIndex, picdexerMappingPayload); err != nil {
		return err
	}
	if err := s.setupIndex(&client, m, syncOnDateIndex, syncOnDateMappingPayload); err != nil {
		return err
	}
	return nil
}

func (s *Setup) setupIndex(client *http.Client, m ESManagerInterface, index string, mapping string) error {
	mappingExists, err := m.MappingAlreadyExist(client, index)
	if err != nil {
		return fmt.Errorf("error while checking if %v mapping already exists: %w", index,  err)
	}
	if mappingExists {
		log.Info().Msgf("Elasticsearch %v mapping already exists, deleting...", index)
		err := m.DeleteMapping(client, index)
		if err != nil {
			return fmt.Errorf("error while deleting %v mapping: %w", index, err)
		}
	}
	if err = m.PutMapping(client, index, mapping); err != nil {
		return fmt.Errorf("error while pushing %v mapping: %w", index, err)
	}
	return nil
}

func (s *Setup) SetupElasticsearch() error {
	log.Info().Msgf("Pushing Elasticsearch mapping...")
	m, err := NewESManager(s.esUrl)
	if err != nil {
		return err
	}
	return s.setupElasticsearch(m)
}

type kibTplVar struct {
	FsUrl string
}

func (s *Setup) SetupKibana() error {
	log.Info().Msgf("Pushing Kibana objects...")

	// parse template
	tpl, err := template.New("tpl").Delims("{{{", "}}}").Parse(kibanaComponentsPayload)
	if err != nil {
		return fmt.Errorf("error while parsing template: %w", err)
	}

	// create multipart
	body := new(bytes.Buffer)
	mpart := multipart.NewWriter(body)
	part, err := mpart.CreateFormFile("file", "kibana.ndjson")
	if err != nil {
		return err
	}

	// resolve template in multipart
	vars := kibTplVar{s.fsUrl}
	if err := tpl.Execute(part, vars); err != nil {
		return fmt.Errorf("error while resolving template: %w", err)
	}
	if err = mpart.Close(); err != nil {
		return err
	}

	// query
	req, err := http.NewRequest(http.MethodPost, s.kibUrl, body)
	if err != nil {
		return fmt.Errorf("error while creating http request: %w", err)
	}
	req.URL.Path = "/api/saved_objects/_import"
	q := req.URL.Query()
	q.Add("overwrite", "true")
	req.URL.RawQuery = q.Encode()
	req.Header.Add("kbn-xsrf", "true")
	req.Header.Add("Content-type", fmt.Sprintf("multipart/form-data; boundary=%s", mpart.Boundary()))
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while pushing mapping: %w", err)
	}

	// check response
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		if err := logReader(resp.Body); err != nil {
			return fmt.Errorf("error while logging response body: %s", err)
		}
		return fmt.Errorf("wrong status code: %d (body content logged)", resp.StatusCode)
	}

	log.Info().Msgf("Kibana objects pushed")
	return nil
}
