package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/barasher/picdexer/internal/metadata"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	esDocIdentifier = "esDocId"
	bulkSuffix      = "_bulk"
	ndJsonMimeType  = "application/x-ndjson"
	baseSyncDate    = 946684800 * 1000 // 2000-01-01
)

type EsDoc struct {
	Header   EsHeader
	Document interface{}
}

type EsHeader struct {
	Index EsHeaderIndex `json:"index"`
}

type EsHeaderIndex struct {
	Index string `json:"_index"`
	ID    string `json:"_id"`
}

type EsChildDoc struct {
	Child      string
	SyncedDate uint64
}

type EsPusher struct {
	bulkSize int
	url      string
	dateSync map[string]uint64
}

type SyncOnDateBody struct {
	Date       uint64
	SyncedDate uint64
	Key        string
	PicId      string
}

func NewEsPusher(bulkSize int, opts ...func(*EsPusher) error) (*EsPusher, error) {
	if bulkSize <= 0 {
		return nil, fmt.Errorf("bulkSize should be >0 (%v)", bulkSize)
	}
	p := &EsPusher{
		bulkSize: bulkSize,
		dateSync: make(map[string]uint64),
	}
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

func SyncOnDate(kw string, d time.Time) func(*EsPusher) error {
	return func(p *EsPusher) error {
		p.dateSync[kw] = uint64(d.Unix() * 1000)
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
					log.Info().Msgf("Pushing ES bulk (%v docs)...", bufferDocCount)
					if err := collectFct(ctx, &buffer); err != nil {
						return fmt.Errorf("error while sinking buffer: %w", err)
					}
				}
				return nil
			}
			if err := jsonEncoder.Encode(doc.Header); err != nil {
				log.Debug().Str(esDocIdentifier, doc.Header.Index.ID).Msgf("Header: %v", doc.Header)
				return fmt.Errorf("error while encoding header: %w", err)
			}
			if err := jsonEncoder.Encode(doc.Document); err != nil {
				log.Debug().Str(esDocIdentifier, doc.Header.Index.ID).Msgf("Body: %v", doc.Document)
				return fmt.Errorf("error while encoding body: %w", err)
			}
			bufferDocCount++
			if bufferDocCount == pusher.bulkSize {
				log.Info().Msgf("Pushing ES bulk (%v docs)...", bufferDocCount)
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
		b, err := io.ReadAll(reader)
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
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error while reading response body: %w", err)
		}
		log.Error().Msgf("Response body: %v", string(b))
		return fmt.Errorf("wrong status code (%v)", resp.StatusCode)
	}

	return nil
}

func (pusher *EsPusher) ConvertMetadataToEsDoc(ctx context.Context, in chan metadata.PictureMetadata, out chan EsDoc) error {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return nil
		case cur, ok := <-in:
			if !ok {
				return nil
			}
			// main doc
			out <- EsDoc{
				Header: EsHeader{
					Index: EsHeaderIndex{
						Index: "picdexer",
						ID:    cur.FileID,
					},
				},
				Document: cur,
			}
			// date sync
			if cur.Date != nil {
				for kw, d := range pusher.dateSync {
					for _, curKw := range cur.Keywords {
						if curKw == kw {
							syncDoc := EsDoc{
								Header: EsHeader{
									Index: EsHeaderIndex{
										Index: "sync-on-date",
										ID:    kw + "_" + cur.FileID,
									},
								},
								Document: SyncOnDateBody{
									Date:       *cur.Date,
									SyncedDate: *cur.Date - d + baseSyncDate,
									Key:        kw,
									PicId:      cur.FileID,
								},
							}
							log.Debug().Msgf("%v matches %v keyword : %v", cur.FileID, kw, syncDoc)
							out <- syncDoc
							break
						}
					}
				}
			}
		}
	}
	return nil
}
