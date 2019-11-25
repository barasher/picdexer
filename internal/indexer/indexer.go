package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	exif "github.com/barasher/go-exiftool"
	"github.com/barasher/picdexer/internal/model"
	"github.com/sirupsen/logrus"
)

const (
	APERTURE_KEY    = "Aperture"
	SHUTTER_KEY     = "ShutterSpeed"
	KEYWORDS_KEY    = "Keywords"
	CAMERA_KEY      = "Model"
	LENS_KEY        = "LensModel"
	MIMETYPE_KEY    = "MIMEType"
	HEIGHT_KEY      = "ImageHeight"
	WIDTH_KEY       = "ImageWidth"
	CAPTUREDATE_KEY = "CreateDate"
	GPS_KEY         = "GPSPosition"
	ISO_KEY         = "ISO"

	SRC_DATE_FORMAT = "2006:01:02 15:04:05"
	IMAGE_MIME_TYPE = "image/"

	NDJSON_CONTENTTYPE = "application/x-ndjson"
)

type Indexer struct {
	input       string
	exif        *exif.Exiftool
	threadCount int
}

type bulkEntryHeader struct {
	Index struct {
		Index string `json:"_index"`
		ID    string `json:"_id"`
	} `json:"index"`
}

func NewIndexer(opts ...func(*Indexer) error) (*Indexer, error) {
	idxer := &Indexer{}
	for _, opt := range opts {
		if err := opt(idxer); err != nil {
			return nil, fmt.Errorf("Initialization error: %v", err)
		}
	}

	et, err := exif.NewExiftool()
	if err != nil {
		return idxer, fmt.Errorf("error while initializing metadata extractor: %v", err)
	}
	idxer.exif = et
	idxer.threadCount = runtime.GOMAXPROCS(runtime.NumCPU())

	return idxer, nil
}

func Input(input string) func(*Indexer) error {
	return func(idxer *Indexer) error {
		idxer.input = input
		return nil
	}
}

func (idxer *Indexer) Close() error {
	if idxer.exif != nil {
		if err := idxer.exif.Close(); err != nil {
			logrus.Errorf("error while closing exiftool: %v", err)
		}
	}
	return nil
}

type extractTask struct {
	path string
	info os.FileInfo
}

type printTask struct {
	header bulkEntryHeader
	pic    model.Model
}

func startPrint(ctx context.Context, cancel context.CancelFunc, globalWg *sync.WaitGroup, printChan chan printTask, writer io.Writer) {
	defer globalWg.Done()
	jsonEncoder := json.NewEncoder(writer)
	for task := range printChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := jsonEncoder.Encode(task.header); err != nil {
			logrus.Errorf("error while encoding header: %v", err)
			cancel()
			return
		}
		if err := jsonEncoder.Encode(task.pic); err != nil {
			logrus.Errorf("error while encoding json: %v", err)
			cancel()
			return
		}
	}
}

func (idxer *Indexer) startConsumers(ctx context.Context, cancel context.CancelFunc, globalWg *sync.WaitGroup, consumeChan chan extractTask, printChan chan printTask) {
	defer globalWg.Done()
	threadCount := idxer.threadCount
	var consumeWg sync.WaitGroup
	consumeWg.Add(threadCount)
	for i := 0; i < threadCount; i++ {
		go func(id int) {
			defer consumeWg.Done()
			for task := range consumeChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				pic, err := idxer.convert(task.path, task.info)
				if err != nil {
					logrus.Errorf("%v: %v", task.path, err)
					cancel()
					return
				} else {
					if pic.MimeType != nil && strings.HasPrefix(*pic.MimeType, IMAGE_MIME_TYPE) {
						header, err := getBulkEntryHeader(task.path, pic)
						if err != nil {
							logrus.Errorf("error while generating header: %v", err)
							cancel()
							return
						}
						printChan <- printTask{header: header, pic: pic}

					}
				}
			}
		}(i)
	}
	consumeWg.Wait()
	close(printChan)
}

func (idxer *Indexer) Dump(writer io.Writer) error {
	ctx, cancel := context.WithCancel(context.Background())

	consumeChan := make(chan extractTask, idxer.threadCount*3)
	printChan := make(chan printTask, idxer.threadCount)
	var wg sync.WaitGroup
	wg.Add(2)

	go startPrint(ctx, cancel, &wg, printChan, writer)
	go idxer.startConsumers(ctx, cancel, &wg, consumeChan, printChan)

	err := filepath.Walk(idxer.input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			consumeChan <- extractTask{
				path: path,
				info: info,
			}
		}
		return nil
	})
	if err != nil {
		cancel()
		return fmt.Errorf("error while browsing directory: %v", err)
	}
	close(consumeChan)
	wg.Wait()
	return nil
}

func (idxer *Indexer) convert(f string, fInfo os.FileInfo) (model.Model, error) {
	logrus.Infof("%v", f)
	pic := model.Model{}

	metas := idxer.exif.ExtractMetadata(f)
	if len(metas) != 1 {
		return pic, fmt.Errorf("wrong metadata count (%v)", len(metas))
	}
	meta := metas[0]

	pic.Aperture = getFloat64(meta, APERTURE_KEY)
	pic.ISO = getInt64(meta, ISO_KEY)
	pic.ShutterSpeed = getString(meta, SHUTTER_KEY)
	pic.CameraModel = getString(meta, CAMERA_KEY)
	pic.LensModel = getString(meta, LENS_KEY)
	pic.MimeType = getString(meta, MIMETYPE_KEY)
	pic.Height = getInt64(meta, HEIGHT_KEY)
	pic.Width = getInt64(meta, WIDTH_KEY)
	pic.Keywords = getStrings(meta, KEYWORDS_KEY)
	pic.FileSize = uint64(fInfo.Size())
	pic.FileName = fInfo.Name()
	pic.Date = getDate(meta, CAPTUREDATE_KEY)
	pic.GPS = getGPS(meta, GPS_KEY)

	components := strings.Split(f, string(os.PathSeparator))
	if len(components) > 1 {
		pic.Folder = components[len(components)-2]
	}

	return pic, nil
}

func (idxer *Indexer) Push(esUrl string, buffer *bytes.Buffer) error {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := httpClient.Post(esUrl, NDJSON_CONTENTTYPE, buffer)
	if err != nil {
		return fmt.Errorf("Error while pushing to Elasticsearch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Wrong status code (%v)", resp.StatusCode)
	}

	return nil
}