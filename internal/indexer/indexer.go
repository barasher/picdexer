package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"encoding/json"

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
	CAPTUREDATE_KEY = "DateTimeCreated"
	GPS_KEY         = "GPSPosition"

	SRC_DATE_FORMAT = "2006:01:02 15:04:05" 

	ES_BULK_LINE_HEADER="{ \"index\":{} }"
)

type Indexer struct {
	input string
	exif  *exif.Exiftool
}

func NewIndexer(opts ...func(*Indexer) error) (*Indexer, error) {
	idxer := &Indexer{}
	for _, opt := range opts {
		if err := opt(idxer); err != nil {
			return nil, fmt.Errorf("Initialization error: %v", err)
		}
	}
	return idxer, nil
}

func Input(input string) func(*Indexer) error {
	return func(idxer *Indexer) error {
		idxer.input = input
		return nil
	}
}

func (idxer *Indexer) Close() error {
	if err := idxer.exif.Close(); err != nil {
		logrus.Errorf("error while closing exiftool: %v", err)
	}
	return nil
}

func (idxer *Indexer) Index() error {
	et, err := exif.NewExiftool()
	if err != nil {
		return fmt.Errorf("error while initializing metadata extractor: %v", err)
	}
	idxer.exif = et

	jsonEncoder := json.NewEncoder(os.Stdout)

	err = filepath.Walk(idxer.input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			pic, err := idxer.convert(path, info)
			if err != nil {
				logrus.Errorf("%v: %v", path, err)
			} else {
				fmt.Fprintln(os.Stdout, ES_BULK_LINE_HEADER)
				if err := jsonEncoder.Encode(pic); err != nil {
					logrus.Errorf("error while encoding json: %v", err)
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

func (*Indexer) getFloat(m exif.FileMetadata, k string) float64 {
	v, found := m.Fields[k]
	if !found {
		return 0
	}
	return v.(float64)
}

func (*Indexer) getString(m exif.FileMetadata, k string) string {
	v, found := m.Fields[k]
	if !found {
		return ""
	}
	return v.(string)
}

func (idxer *Indexer) convert(f string, fInfo os.FileInfo) (model.Model, error) {
	logrus.Infof("%v", f)
	pic := model.Model{}

	metas := idxer.exif.ExtractMetadata(f)
	if len(metas) != 1 {
		return pic, fmt.Errorf("wrong metadata count (%v)", len(metas))
	}
	meta := metas[0]

	pic.Aperture = float32(idxer.getFloat(meta, APERTURE_KEY))
	pic.ShutterSpeed = idxer.getString(meta, SHUTTER_KEY)
	pic.CameraModel = idxer.getString(meta, CAMERA_KEY)
	pic.LensModel = idxer.getString(meta, LENS_KEY)
	pic.MimeType = idxer.getString(meta, MIMETYPE_KEY)
	pic.Height = uint32(idxer.getFloat(meta, HEIGHT_KEY))
	pic.Width = uint32(idxer.getFloat(meta, WIDTH_KEY))
	pic.GPS = meta.Fields[GPS_KEY].(string)                 // enrich
	pic.FileSize = uint32(fInfo.Size())

	rawKws, found := meta.Fields[KEYWORDS_KEY]
	if found {
		pic.Keywords = make([]string, len(rawKws.([]interface{})))
		for i, v := range rawKws.([]interface{}) {
			pic.Keywords[i] = v.(string)
		}
	}

	rawDate, found := meta.Fields[CAPTUREDATE_KEY]
	if found {
		var err error
		if pic.CaptureDate, err = time.Parse(SRC_DATE_FORMAT, rawDate.(string)) ; err != nil {
			return pic, fmt.Errorf("error while parsing date (%v): %v", rawDate.(string), err)
		}
	}

	logrus.Infof("%v", pic)
	return pic, nil
}
