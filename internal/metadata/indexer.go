package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/barasher/picdexer/internal/common"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	exif "github.com/barasher/go-exiftool"
)

const (
	apertureKey    = "Aperture"
	shutterKey     = "ShutterSpeed"
	keywordsKey    = "Keywords"
	cameraKey      = "Model"
	lensKey        = "LensModel"
	mimeTypeKey    = "MIMEType"
	heightKey      = "ImageHeight"
	widthKey       = "ImageWidth"
	captureDateKey = "CreateDate"
	gpsKey         = "GPSPosition"
	isoKey         = "ISO"

	srcDateFormat = "2006:01:02 15:04:05"

	ndJsonMimeType = "application/x-ndjson"
	bulkSuffix     = "_bulk"

	defaultExtrationThreadCount = 4
	defaultToExtractChannelSize = 50
)

type Indexer struct {
	conf conf.ElasticsearchConf
	exif *exif.Exiftool
}

type Collector func(ctx context.Context, cancel context.CancelFunc, collectChan chan printTask) error

type bulkEntryHeader struct {
	Index struct {
		Index string `json:"_index"`
		ID    string `json:"_id"`
	} `json:"index"`
}

func (idxer *Indexer) extractionThreadCount() int {
	n := idxer.conf.ExtractionThreadCount
	if n < 1 {
		n = defaultExtrationThreadCount
	}
	return n
}

func (idxer *Indexer) toExtractChannelSize() int {
	n := idxer.conf.ToExtractChannelSize
	if n < 1 {
		n = defaultToExtractChannelSize
	}
	return n
}

func NewIndexer(c conf.ElasticsearchConf) (*Indexer, error) {
	idxer := &Indexer{conf: c}

	et, err := exif.NewExiftool()
	if err != nil {
		return idxer, fmt.Errorf("error while initializing metadata extractor: %v", err)
	}
	idxer.exif = et

	return idxer, nil
}

func (idxer *Indexer) Close() error {
	if idxer.exif != nil {
		if err := idxer.exif.Close(); err != nil {
			log.Error().Msgf("error while closing exiftool: %v", err)
		}
	}
	return nil
}

type ExtractTask struct {
	Path string
	Info os.FileInfo
}

type printTask struct {
	header bulkEntryHeader
	pic    Model
}

func (idxer *Indexer) convert(ctx context.Context, f string, fInfo os.FileInfo) (Model, error) {
	log.Info().Str(logFileIdentifier, f).Msg("Converting...")
	pic := Model{}

	metas := idxer.exif.ExtractMetadata(f)
	if len(metas) != 1 {
		return pic, fmt.Errorf("wrong metadata count (%v)", len(metas))
	}
	meta := metas[0]

	pic.ImportID = common.GetImportID(ctx)
	pic.Aperture = getFloat64(meta, apertureKey)
	pic.ISO = getInt64(meta, isoKey)
	pic.ShutterSpeed = getString(meta, shutterKey)
	pic.CameraModel = getString(meta, cameraKey)
	pic.LensModel = getString(meta, lensKey)
	pic.MimeType = getString(meta, mimeTypeKey)
	pic.Height = getInt64(meta, heightKey)
	pic.Width = getInt64(meta, widthKey)
	pic.Keywords = getStrings(meta, keywordsKey)
	pic.FileSize = uint64(fInfo.Size())
	pic.FileName = fInfo.Name()
	pic.Date = getDate(meta, captureDateKey)
	pic.GPS = getGPS(meta, gpsKey)

	components := strings.Split(f, string(os.PathSeparator))
	if len(components) > 1 {
		pic.Folder = components[len(components)-2]
	}

	return pic, nil
}

func (idxer *Indexer) ExtractFolder(ctx context.Context, root string) error {
	return idxer.extractFolder(ctx, root, idxer.collectToPrint)
}

func (idxer *Indexer) ExtractAndPushFolder(ctx context.Context, root string) error {
	return idxer.extractFolder(ctx, root, idxer.collectToPush)
}

func (idxer *Indexer) extractFolder(ctx context.Context, root string, collector Collector) error {
	ctx, cancel := context.WithCancel(ctx)
	toExtractChan := make(chan ExtractTask, idxer.toExtractChannelSize())
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		idxer.extractTasks(ctx, cancel, toExtractChan, collector)
	}()

	err := common.BrowseImages(root, func(path string, info os.FileInfo) {
		toExtractChan <- ExtractTask{
			Path: path,
			Info: info,
		}
	})
	close(toExtractChan)
	if err != nil {
		cancel()
		return fmt.Errorf("error while browsing directory: %v", err)
	}

	wg.Wait()
	return nil
}

func (idxer *Indexer) ExtractTasks(ctx context.Context, tasks <-chan ExtractTask) error {
	ctx, cancel := context.WithCancel(ctx)
	return idxer.extractTasks(ctx, cancel, tasks, idxer.collectToPrint)
}

func (idxer *Indexer) ExtractAndPushTasks(ctx context.Context, tasks <-chan ExtractTask) error {
	ctx, cancel := context.WithCancel(ctx)
	return idxer.extractTasks(ctx, cancel, tasks, idxer.collectToPush)

}

func (idxer *Indexer) extractTasks(ctx context.Context, cancel context.CancelFunc, tasks <-chan ExtractTask, collector Collector) error {
	extractionThreadCount := idxer.extractionThreadCount()
	wgExtract := sync.WaitGroup{}
	wgExtract.Add(extractionThreadCount)
	wgCollect := sync.WaitGroup{}
	wgCollect.Add(1)

	toDumpChan := make(chan printTask, idxer.extractionThreadCount()+1)
	go func() {
		defer wgCollect.Done()
		collector(ctx, cancel, toDumpChan)
	}()

	for i := 0; i < extractionThreadCount; i++ {
		go func(goRoutineId int) {
			defer wgExtract.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-tasks:
					if !ok {
						return
					}
					pic, err := idxer.convert(ctx, task.Path, task.Info)
					if err != nil {
						log.Error().Str(logFileIdentifier, task.Path).Msgf("conversion error: %v", err)
						cancel()
						return
					} else {
						header, err := getBulkEntryHeader(task.Path, pic)
						if err != nil {
							log.Error().Str(logFileIdentifier, task.Path).Msgf("error while generating header: %v", err)
							cancel()
							return
						}
						toDumpChan <- printTask{header: header, pic: pic}
					}
				}
			}
		}(i)
	}

	wgExtract.Wait()
	close(toDumpChan)

	wgCollect.Wait()

	return nil
}

func (indexer *Indexer) sinkCollectChan(ctx context.Context, cancel context.CancelFunc, collectChan chan printTask, writer io.Writer) error {
	jsonEncoder := json.NewEncoder(writer)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("collectToPrint cancelled")
		case task, ok := <-collectChan:
			if !ok {
				return nil
			}
			if err := jsonEncoder.Encode(task.header); err != nil {
				cancel()
				return fmt.Errorf("error while encoding header: %w", err)
			}
			if err := jsonEncoder.Encode(task.pic); err != nil {
				cancel()
				return fmt.Errorf("error while encoding json: %v", err)
			}
		}
	}
}

func (indexer *Indexer) collectToPrint(ctx context.Context, cancel context.CancelFunc, collectChan chan printTask) error {
	return indexer.sinkCollectChan(ctx, cancel, collectChan, os.Stdout)
}

func (idxer *Indexer) collectToPush(ctx context.Context, cancel context.CancelFunc, collectChan chan printTask) error {
	buffer := bytes.Buffer{}
	if err := idxer.sinkCollectChan(ctx, cancel, collectChan, &buffer); err != nil {
		return fmt.Errorf("error while sinking collecting channel: %w", err)
	}

	log.Info().Msgf("Pushing to Elasticsearch...")
	u, err := url.Parse(idxer.conf.Url)
	if err != nil {
		return fmt.Errorf("error while parsing elasticsearch url (%v): %w", idxer.conf.Url, err)
	}
	u.Path = path.Join(u.Path, bulkSuffix)

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := httpClient.Post(u.String(), ndJsonMimeType, &buffer)
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
