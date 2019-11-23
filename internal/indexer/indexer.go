package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	SRC_DATE_FORMAT     = "2006:01:02 15:04:05"
	ES_BULK_LINE_HEADER = "{ \"index\":{} }"
	IMAGE_MIME_TYPE     = "image/"
)

type Indexer struct {
	input string
	exif  *exif.Exiftool
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

func (idxer *Indexer) Index() error {
	jsonEncoder := json.NewEncoder(os.Stdout)
	err := filepath.Walk(idxer.input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			pic, err := idxer.convert(path, info)
			if err != nil {
				logrus.Errorf("%v: %v", path, err)
			} else {
				if pic.MimeType != nil && strings.HasPrefix(*pic.MimeType, IMAGE_MIME_TYPE) {
					//header
					header, err := getBulkEntryHeader(path, pic)
					if err != nil {
						logrus.Errorf("error while generating header: %v", err)
						return nil
					}
					if err := jsonEncoder.Encode(header); err != nil {
						logrus.Errorf("error while encoding header: %v", err)
						return nil
					}
					// body
					if err := jsonEncoder.Encode(pic); err != nil {
						logrus.Errorf("error while encoding json: %v", err)
						return nil
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while browsing directory: %v", err)
	}

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
