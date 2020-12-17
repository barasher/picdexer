package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	esDocIdentifier = "esDocId"
	defaultBulkSize = 30
	bulkSuffix      = "_bulk"
	ndJsonMimeType  = "application/x-ndjson"
)

type EsDoc struct {
	Header struct {
		Index string `json:"_index"`
		ID    string `json:"_id"`
	} `json:"index"`
	Document interface{}
}

type EsPusher struct {
	bulkSize int
	url      string
}

func NewEsPusher(opts ...func(*EsPusher) error) (*EsPusher, error) {
	p := &EsPusher{bulkSize: defaultBulkSize}
	for _, cur := range opts {
		if err := cur(p); err != nil {
			return nil, fmt.Errorf("error while creating EsPusher: %w", err)
		}
	}
	return p, nil
}

func BulkSize(size int) func(*EsPusher) error {
	return func(p *EsPusher) error {
		if size <= 0 {
			return fmt.Errorf("wrong bulksize value (%v), must be > 0", size)
		}
		p.bulkSize = size
		return nil
	}
}

func EsUrl(url string) func(*EsPusher) error {
	return func(p *EsPusher) error {
		p.url = url
		return nil
	}
}

func (pusher *EsPusher) sinkChan(ctx context.Context, inEsDocChan chan EsDoc, collectFct func(ctx context.Context, reader io.Reader) error) error {
	buffer := bytes.Buffer{}
	jsonEncoder := json.NewEncoder(&buffer)
	bufferDocCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case doc, ok := <-inEsDocChan:
			if !ok {
				if bufferDocCount > 0 {
					if err := collectFct(ctx, &buffer); err != nil {
						return fmt.Errorf("error while sinking buffer: %w", err)
					}
				}
				return nil
			}
			if err := jsonEncoder.Encode(doc.Header); err != nil {
				log.Debug().Str(esDocIdentifier, doc.Header.ID).Msgf("Header: %v", doc.Header)
				return fmt.Errorf("error while encoding header: %w", err)
			}
			if err := jsonEncoder.Encode(doc.Document); err != nil {
				log.Debug().Str(esDocIdentifier, doc.Header.ID).Msgf("Body: %v", doc.Document)
				return fmt.Errorf("error while encoding body: %w", err)
			}
			bufferDocCount++
			if bufferDocCount == pusher.bulkSize {
				if err := collectFct(ctx, &buffer); err != nil {
					return fmt.Errorf("error while sinking buffer: %w", err)
				}
				bufferDocCount = 0
			}
		}
	}

	return nil
}

func (pusher *EsPusher) Print(ctx context.Context, inEsDocChan chan EsDoc) error {
	return pusher.sinkChan(ctx, inEsDocChan, func(ctx context.Context, reader io.Reader) error {
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("error while printing documents: %w", err)
		}
		fmt.Printf("%v", string(b))
		return nil
	})
}

func (pusher *EsPusher) Push(ctx context.Context, inEsDocChan chan EsDoc) error {
	return pusher.sinkChan(ctx, inEsDocChan, func(ctx context.Context, reader io.Reader) error {
		return pusher.pushToEs(ctx, reader)
	})
}

func (pusher *EsPusher) pushToEs(ctx context.Context, body io.Reader) error {
	u, err := url.Parse(pusher.url)
	if err != nil {
		return fmt.Errorf("error while parsing elasticsearch url (%v): %w", pusher.url, err)
	}
	u.Path = path.Join(u.Path, bulkSuffix)

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := httpClient.Post(u.String(), ndJsonMimeType, body)
	if err != nil {
		return fmt.Errorf("error while pushing to Elasticsearch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error while reading response body: %w", err)
		}
		log.Error().Msgf("Response body: %v", string(b))
		return fmt.Errorf("wrong status code (%v)", resp.StatusCode)
	}

	return nil
}
