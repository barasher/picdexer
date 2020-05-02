package setup

import (
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/rakyll/statik/fs"
	_ "github.com/barasher/picdexer/internal/setup/statik"
	"net/http"
)

const mappingPath = "/picdexer"


type ESManager struct {
	conf conf.ElasticsearchConf
	fs   http.FileSystem
}

func NewESManager(c conf.ElasticsearchConf) (*ESManager, error) {
	var err error
	m := &ESManager{conf: c}
	if m.fs, err = fs.New(); err != nil {
		return nil, fmt.Errorf("error while loading fs: %w", err)
	}
	return m, nil
}

func (s *ESManager) simpleMappingQuery(client *http.Client, method string, expStatus int) error {
	req, err := http.NewRequest(method, s.conf.Url, nil)
	if err != nil {
		return fmt.Errorf("error while creating http request: %w", err)
	}
	req.URL.Path = mappingPath
	resp, err := client.Do(req)
	if err != nil {
		return  fmt.Errorf("error while executing http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expStatus {
		if err := logReader(resp.Body); err != nil {
			return  fmt.Errorf("error while logging response body: %s", err)
		}
		return  fmt.Errorf("unexpected status code (%v), body logged", resp.StatusCode)
	}
	return  nil
}

func (s *ESManager) MappingAlreadyExist(client *http.Client) (bool, error) {
	err :=  s.simpleMappingQuery(client, http.MethodGet, http.StatusOK)
	return err == nil, err
}

func (s *ESManager) DeleteMapping(client *http.Client) error {
	return  s.simpleMappingQuery(client, http.MethodDelete, http.StatusOK)
}

func (s *ESManager) PutMapping(client *http.Client) error {
	r, err := s.fs.Open("/mapping.json")
	if err != nil {
		return fmt.Errorf("error while reading mapping: %w", err)
	}
	defer r.Close()

	req, err := http.NewRequest(http.MethodPut, s.conf.Url, r)
	if err != nil {
		return fmt.Errorf("error while creating http request: %w", err)
	}
	req.URL.Path = mappingPath
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while pushing mapping: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		if err := logReader(resp.Body); err != nil {
			return fmt.Errorf("error while logging response body: %s", err)
		}
		return fmt.Errorf("wrong status code: %d (body content logged)", resp.StatusCode)
	}

	return nil
}