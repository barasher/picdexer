package internal

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/barasher/picdexer/internal/common"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

	defaultExtrationThreadCount = 4

)

var defaultDate = uint64(0)

type PictureMetadata struct {
	FileName     string
	Folder       string
	ImportID     string
	FileSize     uint64
	ISO          *uint64    `json:",omitempty"`
	Aperture     *float64   `json:",omitempty"`
	ShutterSpeed *string    `json:",omitempty"`
	Keywords     []string   `json:",omitempty"`
	CameraModel  *string    `json:",omitempty"`
	LensModel    *string    `json:",omitempty"`
	MimeType     *string    `json:",omitempty"`
	Height       *uint64    `json:",omitempty"`
	Width        *uint64    `json:",omitempty"`
	Date         *uint64    `json:",omitempty"`
	ParsedDate   *time.Time `json:"-"`
	GPS          *string    `json:",omitempty"`
}

type MetadataExtractor struct {
	threadCount int
	exif *exif.Exiftool
}

func NewMetadataExtractor(opts ...func(*MetadataExtractor) error) (*MetadataExtractor, error) {
	e := &MetadataExtractor{threadCount:defaultExtrationThreadCount}

	et, err := exif.NewExiftool()
	if err != nil {
		return nil, fmt.Errorf("error while initializing Exiftool: %v", err)
	}
	e.exif = et

	for _, cur := range opts {
		if err := cur(e); err != nil {
			return nil, fmt.Errorf("error while creating MetadataExtractor: %w", err)
		}
	}
	return e, nil
}

func MetadataExtractorThreadCount(c int) func(*MetadataExtractor) error {
	return func(e *MetadataExtractor) error {
		if c <= 0 {
			return fmt.Errorf("wrong thread count value (%v), must be > 0", c)
		}
		e.threadCount = c
		return nil
	}
}

func (ext *MetadataExtractor) Close() error {
	if ext.exif != nil {
		if err := ext.exif.Close(); err != nil {
			log.Error().Msgf("error while closing exiftool: %v", err)
		}
	}
	return nil
}

func (ext *MetadataExtractor) ExtractMetadata(ctx context.Context, inTaskChan chan Task, outPicMetaChan chan PictureMetadata) error {
	wg := sync.WaitGroup{}
	wg.Add(ext.threadCount)

	for i := 0; i < ext.threadCount; i++ {
		go func(goRoutineId int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-inTaskChan:
					if !ok {
						return
					}
					picMeta, err := ext.extractMetadataFromFile(ctx, task.Path, task.Info)
					if err != nil {
						log.Error().Str(logFileIdentifier, task.Path).Msgf("conversion error: %v", err)
					} else {
						outPicMetaChan <- picMeta
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(outPicMetaChan)

	return nil
}

func (ext *MetadataExtractor) extractMetadataFromFile(ctx context.Context, f string, fInfo os.FileInfo) (PictureMetadata, error) {
	log.Info().Str(logFileIdentifier, f).Msg("Extracting metadata...")
	pic := PictureMetadata{}

	metas := ext.exif.ExtractMetadata(f)
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
		log.Warn().Str(logFileIdentifier, m.File).Msgf("error while extracting key %v as string: %v", k, err)
	}
	return nil
}

func getStrings(m exif.FileMetadata, k string) []string {
	v, err := m.GetStrings(k)
	switch {
	case err == nil:
		return v
	case !errors.Is(err, exif.ErrKeyNotFound):
		log.Warn().Str(logFileIdentifier, m.File).Msgf("error while extracting key %v as string slice: %v", k, err)
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
		log.Warn().Str(logFileIdentifier, m.File).Msgf("error while extracting key %v as int: %v", k, err)
	}
	return nil
}

func getFloat64(m exif.FileMetadata, k string) *float64 {
	v, err := m.GetFloat(k)
	switch {
	case err == nil:
		return &v
	case !errors.Is(err, exif.ErrKeyNotFound):
		log.Warn().Str(logFileIdentifier, m.File).Msgf("error while extracting key %v as float: %v", k, err)
	}
	return nil
}

func getDate(m exif.FileMetadata, k string) *uint64 {
	if strDate := getString(m, k); strDate != nil {
		if d, err := time.Parse(srcDateFormat, *strDate); err != nil {
			log.Warn().Str(logFileIdentifier, m.File).Msgf("error while parsing date from field %v (%v): %v", k, *strDate, err)
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
			log.Warn().Str(logFileIdentifier, m.File).Msgf("error while parsing GPS coordinates from field %v (%v): %v", k, *rawGPS, err)
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

