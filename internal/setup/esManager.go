package setup

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"
)

const mappingPath = "/picdexer"

//go:embed assets/mapping.json
var esMappingPayload string

type ESManager struct {
	url string
}

func NewESManager(url string) (*ESManager, error) {
	return &ESManager{url: url}, nil
}

func (s *ESManager) simpleMappingQuery(client *http.Client, method string) (int, error) {
	req, err := http.NewRequest(method, s.url, nil)
	if err != nil {
		return -1, fmt.Errorf("error while creating http request: %w", err)
	}
	req.URL.Path = mappingPath
	resp, err := client.Do(req)
	if err != nil {
		return -1, fmt.Errorf("error while executing http request: %w", err)
	}
	defer resp.Body.Close()
	if err := logReader(resp.Body); err != nil {
		return -1, fmt.Errorf("error while logging response body: %s", err)
	}

	return resp.StatusCode, nil
}

func (s *ESManager) MappingAlreadyExist(client *http.Client) (bool, error) {
	status, err := s.simpleMappingQuery(client, http.MethodGet)
	switch {
	case err != nil:
		return false, err
	case status == http.StatusOK:
		return true, nil
	case status == http.StatusNotFound:
		return false, nil
	default:
		return true, fmt.Errorf("unexpected status code (%v)", status)
	}
}

func (s *ESManager) DeleteMapping(client *http.Client) error {
	status, err := s.simpleMappingQuery(client, http.MethodDelete)
	switch {
	case err != nil:
		return err
	case status == http.StatusOK:
		return nil
	default:
		return fmt.Errorf("unexpected status code (%v)", status)
	}
}

func (s *ESManager) PutMapping(client *http.Client) error {
	req, err := http.NewRequest(http.MethodPut, s.url, strings.NewReader(esMappingPayload))
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
