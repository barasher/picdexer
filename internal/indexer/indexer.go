package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"strings"

	"strconv"

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

	ES_BULK_LINE_HEADER = "{ \"index\":{} }"

	DECIMAL_PATTERN = "[0-9]{1,}[\\.[0-9]{1,}]{0,1}"
	LAT_PATTERN     = "(" + DECIMAL_PATTERN + ") deg (" + DECIMAL_PATTERN + ")' (" + DECIMAL_PATTERN + ")\" (N|S)"
	LONG_PATTERN    = "(" + DECIMAL_PATTERN + ") deg (" + DECIMAL_PATTERN + ")' (" + DECIMAL_PATTERN + ")\" (E|W)"
	GPS_PATTERN     = LAT_PATTERN + ", " + LONG_PATTERN
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

func getFloat(m exif.FileMetadata, k string) float64 {
	v, found := m.Fields[k]
	if !found {
		return 0
	}
	return v.(float64)
}

func getString(m exif.FileMetadata, k string) string {
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

	pic.Aperture = float32(getFloat(meta, APERTURE_KEY))
	pic.ShutterSpeed = getString(meta, SHUTTER_KEY)
	pic.CameraModel = getString(meta, CAMERA_KEY)
	pic.LensModel = getString(meta, LENS_KEY)
	pic.MimeType = getString(meta, MIMETYPE_KEY)
	pic.Height = uint32(getFloat(meta, HEIGHT_KEY))
	pic.Width = uint32(getFloat(meta, WIDTH_KEY))
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
		if pic.CaptureDate, err = time.Parse(SRC_DATE_FORMAT, rawDate.(string)); err != nil {
			return pic, fmt.Errorf("error while parsing date (%v): %v", rawDate.(string), err)
		}
	}


	if gpsVal, found := meta.Fields[GPS_KEY]; found {
		lat, long, err := convertGPSCoordinates(gpsVal.(string))
		if err != nil {
			return pic, fmt.Errorf("error while converting gps coordinates (%v): %v", gpsVal, err)
		}
		pic.GPS = fmt.Sprintf("%v,%v", lat, long)
	}

	logrus.Infof("%v", pic)
	return pic, nil
}

func degMinSecToDecimal(deg, min, sec, let string) (float32, error) {
	var fDeg, fMin, fSec float64
	var err error
	if fDeg, err = strconv.ParseFloat(deg, 32); err != nil {
		return 0, fmt.Errorf("error while parsing deg %v as float", deg)
	}
	if fMin, err = strconv.ParseFloat(min, 32); err != nil {
		return 0, fmt.Errorf("error while parsing min %v as float", min)
	}
	if fSec, err = strconv.ParseFloat(sec, 32); err != nil {
		return 0, fmt.Errorf("error while parsing sec %v as float", sec)
	}
	var mult float64
	switch {
	case let == "S" || let == "W":
		mult = -1
	case let == "N" || let == "E":
		mult = 1
	default:
		return 0, fmt.Errorf("Unsupported letter (%v)", let)
	}
	return float32((fDeg + fMin/60 + fSec/3600) * mult), nil
}

func skipLastChar(src string) string {
	return src[:len(src)-1]
}

func convertGPSCoordinates(latLong string) (float32, float32, error) {
	sub := strings.Split(latLong, " ")
	if len(sub) != 10 {
		return 0, 0, fmt.Errorf("Parsing inconsistency (%v): %v elements parsed", latLong, len(sub))
	} 
	lat, err := degMinSecToDecimal(sub[0], skipLastChar(sub[2]), skipLastChar(sub[3]), skipLastChar(sub[4]))
	if err != nil {
		return 0,0,fmt.Errorf("error while converting latitude (%v): %v", latLong, err)
	}
	long, err := degMinSecToDecimal(sub[5], skipLastChar(sub[7]), skipLastChar(sub[8]), sub[9])
	if err != nil {
		return 0,0,fmt.Errorf("error while converting longitude (%v): %v", latLong, err)
	}
	return lat, long, nil
}
