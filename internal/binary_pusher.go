package internal

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type pusherInterface interface {
	push(bin string, key string) error
}

type pusher struct {
	url string
	httpClient *http.Client
}

func NewPusher(url string) pusher {
	p := pusher{
		url:url,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	return p
}

func (p pusher) push(f string, key string) error {
	body := new(bytes.Buffer)
	mpart := multipart.NewWriter(body)
	part, err := mpart.CreateFormFile("file", key)
	if err != nil {
		return err
	}

	input, err := os.Open(f)
	if err != nil {
		return err
	}
	defer input.Close()
	if _, err := io.Copy(part, input); err != nil {
		return err
	}
	if err = mpart.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", p.url, body)
	if err != nil {
		return err
	}
	req.URL.Path = fmt.Sprintf("/key/%s", key)
	req.Header.Set("Content-type", fmt.Sprintf("multipart/form-data; boundary=%s", mpart.Boundary()))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Unexpected http status (%v)", resp.StatusCode)
	}
	return nil
}

type nopPusher struct{}

func NewNopPusher() nopPusher {
	return nopPusher{}
}

func (nopPusher) push(f string, key string) error {
	return nil
}
