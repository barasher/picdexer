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
	bulkSuffix      = "_bulk"
	ndJsonMimeType  = "application/x-ndjson"
)

type EsDoc struct {
	Header   EsHeader
	Document interface{}
}

type EsHeader struct {
	Index string `json:"_index"`
	ID    string `json:"_id"`
}

type EsPusher struct {
	bulkSize int
	url      string
}

func NewEsPusher(bulkSize int, opts ...func(*EsPusher) error) (*EsPusher, error) {
	if bulkSize <= 0 {
		return nil, fmt.Errorf("bulkSize should be >0 (%v)", bulkSize)
	}
	p := &EsPusher{bulkSize: bulkSize}
	for _, cur := range opts {
		if err := cur(p); err != nil {
			return nil, fmt.Errorf("error while creating EsPusher: %w", err)
		}
	}
	return p, nil
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

func ConvertMetadataToEsDoc(ctx context.Context, in chan PictureMetadata, out chan EsDoc) error {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return nil
		case cur, ok := <-in:
			if !ok {
				return nil
			}
			id, err := getID(cur.SourceFile)
			if err != nil {
				log.Error().Str(logFileIdentifier, cur.SourceFile).Msgf("Error while building document Id: %v", err)
				continue
			}
			out <- EsDoc{
				Header: EsHeader{
					Index: "picdexer",
					ID:    id,
				},
				Document: cur,
			}
		}
	}
	return nil
}
