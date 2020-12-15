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
	conf EsPusherConf
}

type EsPusherConf struct {
	BulkSize int    `json:"bulkSize"`
	Url      string `json:"url"`
}

func NewEsPusher(conf EsPusherConf) (*EsPusher, error) {
	p := EsPusher{conf: conf}
	return &p, nil
}

func (p *EsPusher) bulkSize() int {
	n := p.conf.BulkSize
	if n < 1 {
		n = defaultBulkSize
	}
	return n
}

func (pusher *EsPusher) Push(ctx context.Context, inEsDocChan chan EsDoc) error {
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
					if err := pusher.pushToEs(ctx, &buffer); err != nil {
						return fmt.Errorf("error while pushing to elasticsearch: %w", err)
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
			if bufferDocCount == pusher.bulkSize() {
				if err := pusher.pushToEs(ctx, &buffer); err != nil {
					return fmt.Errorf("error while pushing to elasticsearch: %w", err)
				}
				bufferDocCount = 0
			}
		}
	}

	return nil
}

func (pusher *EsPusher) pushToEs(ctx context.Context, body io.Reader) error {
	u, err := url.Parse(pusher.conf.Url)
	if err != nil {
		return fmt.Errorf("error while parsing elasticsearch url (%v): %w", pusher.conf.Url, err)
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

func (*EsPusher) Print(ctx context.Context, inEsDocChan chan EsDoc) error {
	return nil
}
