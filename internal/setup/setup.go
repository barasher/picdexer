//go:generate echo "Embedding assets..."
//go:generate statik -src assets/ -f

package setup

import (
	"bytes"
	"fmt"
	_ "github.com/barasher/picdexer/internal/setup/statik"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"text/template"
	"time"
)

type ESManagerInterface interface {
	MappingAlreadyExist(client *http.Client) (bool, error)
	DeleteMapping(client *http.Client) error
	PutMapping(client *http.Client) error
}

type Setup struct {
	esUrl  string
	kibUrl string
	fsUrl  string
	fs     http.FileSystem
}

func logReader(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error while reading response body: %w", err)
	}
	log.Debug().Msgf("Response body: %s", string(b))
	return nil
}

func NewSetup(esUrl string, kibUrl string, fsUrl string) (*Setup, error) {
	var err error
	s := &Setup{esUrl: esUrl, kibUrl: kibUrl, fsUrl: fsUrl}
	if s.fs, err = fs.New(); err != nil {
		return nil, fmt.Errorf("error while loading fs: %w", err)
	}
	return s, nil
}

func (s *Setup) setupElasticsearch(m ESManagerInterface) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	mappingExists, err := m.MappingAlreadyExist(&client)
	if err != nil {
		return fmt.Errorf("error while checking if mapping already exists: %w", err)
	}
	if mappingExists {
		log.Info().Msgf("Elasticsearch mapping already exists, deleting...")
		err := m.DeleteMapping(&client)
		if err != nil {
			return fmt.Errorf("error while deleting mapping: %w", err)
		}
	}
	if err = m.PutMapping(&client); err != nil {
		return fmt.Errorf("error while pushing mapping: %w", err)
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
	r, err := s.fs.Open("/kibana.ndjson")
	if err != nil {
		return fmt.Errorf("error while reading kibana saved objects: %w", err)
	}
	defer r.Close()
	strTpl, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error while reading template: %w", err)
	}
	tpl, err := template.New("tpl").Delims("{{{", "}}}").Parse(string(strTpl))
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
