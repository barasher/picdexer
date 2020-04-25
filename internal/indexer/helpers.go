package indexer

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	exif "github.com/barasher/go-exiftool"
	"github.com/sirupsen/logrus"
)

const (
	indexNameNoDate = "picdexerNoDate"
	indexName       = "picdexer"
	importIdCtxKey  = "impID"
)

var defaultDate = uint64(0)

func getID(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("Error while calculating ID for %v: %w", file, err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("Error while calculating ID for %v: %w", file, err)
	}

	return hex.EncodeToString(h.Sum(nil)) + "_" + filepath.Base(file), nil
}

func getString(m exif.FileMetadata, k string) *string {
	v, err := m.GetString(k)
	switch {
	case err == nil:
		return &v
	case !errors.Is(err, exif.ErrKeyNotFound):
		logrus.Warnf("error while extracting key %v as string from %v: %v", k, m.File, err)
	}
	return nil
}

func getStrings(m exif.FileMetadata, k string) []string {
	v, err := m.GetStrings(k)
	switch {
	case err == nil:
		return v
	case !errors.Is(err, exif.ErrKeyNotFound):
		logrus.Warnf("error while extracting key %v as string slice from %v: %v", k, m.File, err)
	}
	return nil
}

func getInt64(m exif.FileMetadata, k string) *uint64 {
	v, err := m.GetInt(k)
	switch {
	case err == nil:
		uv := uint64(v)
		return &uv
	case !errors.Is(err, exif.ErrKeyNotFound):
		logrus.Warnf("error while extracting key %v as int from %v: %v", k, m.File, err)
	}
	return nil
}

func getFloat64(m exif.FileMetadata, k string) *float64 {
	v, err := m.GetFloat(k)
	switch {
	case err == nil:
		return &v
	case !errors.Is(err, exif.ErrKeyNotFound):
		logrus.Warnf("error while extracting key %v as float from %v: %v", k, m.File, err)
	}
	return nil
}

func getDate(m exif.FileMetadata, k string) *uint64 {
	if strDate := getString(m, k); strDate != nil {
		if d, err := time.Parse(srcDateFormat, *strDate); err != nil {
			logrus.Warnf("error while parsing date from field %v (%v) from %v: %v", k, *strDate, m.File, err)
			return &defaultDate
		} else {
			d2 := uint64(d.Unix() * 1000)
			return &d2
		}
	}
	return &defaultDate
}

func getGPS(m exif.FileMetadata, k string) *string {
	if rawGPS := getString(m, k); rawGPS != nil {
		lat, long, err := convertGPSCoordinates(*rawGPS)
		if err != nil {
			logrus.Warnf("error while parsing GPS coordinates from field %v (%v) from %v: %v", k, *rawGPS, m.File, err)
			return nil
		}
		gps := fmt.Sprintf("%v,%v", lat, long)
		return &gps
	}
	return nil
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
		return 0, 0, fmt.Errorf("error while converting latitude (%v): %v", latLong, err)
	}
	long, err := degMinSecToDecimal(sub[5], skipLastChar(sub[7]), skipLastChar(sub[8]), sub[9])
	if err != nil {
		return 0, 0, fmt.Errorf("error while converting longitude (%v): %v", latLong, err)
	}
	return lat, long, nil
}

func getBulkEntryHeader(path string, m Model) (bulkEntryHeader, error) {
	h := bulkEntryHeader{}
	h.Index.Index = indexNameNoDate
	var err error
	if h.Index.ID, err = getID(path); err != nil {
		return h, fmt.Errorf("Error while computing ID: %w", err)
	}
	if m.Date != nil {
		h.Index.Index = indexName
	}
	return h, nil
}

func BuildContext(impID string) context.Context {
	id := impID
	if id == "" {
		id = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return context.WithValue(context.Background(), importIdCtxKey, id)
}

func getImportID(ctx context.Context) string {
	if v := ctx.Value(importIdCtxKey); v != nil {
		return v.(string)
	}
	return ""
}
