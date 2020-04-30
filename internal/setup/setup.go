//go:generate echo "Embedding assets..."
//go:generate statik -src assets/ -f

package setup

import (
	"fmt"
	"github.com/barasher/picdexer/conf"
	_ "github.com/barasher/picdexer/internal/setup/statik"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog/log"
	"io/ioutil"
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
