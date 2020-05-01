//go:generate echo "Embedding assets..."
//go:generate statik -src assets/ -f

package setup

import (
	"bytes"
	"fmt"
	"github.com/barasher/picdexer/conf"
	_ "github.com/barasher/picdexer/internal/setup/statik"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

type Setup struct {
	conf conf.Conf
	fs   http.FileSystem
}

func NewSetup(c conf.Conf) (*Setup, error) {
	var err error
	s := &Setup{conf: c}
	if s.fs, err = fs.New(); err != nil {
		return nil, fmt.Errorf("error while loading fs: %w", err)
	}
	return s, nil
}

func (s *Setup) SetupElasticsearch() error {
	log.Info().Msgf("Pushing Elasticsearch mapping...")
	r, err := s.fs.Open("/mapping.json")
	if err != nil {
		return fmt.Errorf("error while reading mapping: %w", err)
	}
	defer r.Close()

	req, err := http.NewRequest(http.MethodPut, s.conf.Elasticsearch.Url, r)
	req.URL.Path = "/picdexer"
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while pushing mapping: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer req.Body.Close()
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("error while reading response body: %w", err)
		}
		log.Error().Msgf("Response body: %s", string(b))
		return fmt.Errorf("wrong status code: %d (body content logged)", resp.StatusCode)
	}

	log.Info().Msgf("Elasticsearch mapping pushed")
	return nil
}

func (s *Setup) SetupKibana() error {
	log.Info().Msgf("Pushing Kibana objects...")

	r, err := s.fs.Open("/kibana.ndjson")
	if err != nil {
		return fmt.Errorf("error while reading kibana saved objects: %w", err)
	}
	defer r.Close()

	body := new(bytes.Buffer)
	mpart := multipart.NewWriter(body)
	part, err := mpart.CreateFormFile("file", "kibana.ndjson")
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, r); err != nil {
		return err
	}
	if err = mpart.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.conf.Kibana.Url, body)
	req.URL.Path = "/api/saved_objects/_import"
	req.Header.Add("kbn-xsrf", "true")
	req.Header.Add("Content-type", fmt.Sprintf("multipart/form-data; boundary=%s", mpart.Boundary()))
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while pushing mapping: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer req.Body.Close()
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("error while reading response body: %w", err)
		}
		log.Error().Msgf("Response body: %s", string(b))
		return fmt.Errorf("wrong status code: %d (body content logged)", resp.StatusCode)
	}

	log.Info().Msgf("Kibana objects pushed")
	return nil
}
