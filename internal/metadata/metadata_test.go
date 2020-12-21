package metadata

import (
	"context"
	"fmt"
	exif "github.com/barasher/go-exiftool"
	"github.com/barasher/picdexer/internal/browse"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestDegMinSecToDecimal(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inDeg  string
		inMin  string
		inSec  string
		inLet  string
		expOk  bool
		expVal float32
	}{
		{"Unparsable deg", "bla", "1.0", "1.0", "N", false, 0},
		{"Unparsable min", "1.0", "bla", "1.0", "N", false, 0},
		{"Unparsable sec", "1.0", "1.0", "bla", "N", false, 0},
		{"Unsupported letter", "1.0", "1.0", "1.0", "Q", false, 0},
		{"Nominal decimal", "1.0", "11.0", "60.0", "N", true, 1.2},
		{"Nominal integer N", "1", "11", "60", "N", true, 1.2},
		{"Nominal integer E", "1", "11", "60", "E", true, 1.2},
		{"Nominal integer S", "1", "11", "60", "S", true, -1.2},
		{"Nominal integer W", "1", "11", "60", "W", true, -1.2},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			v, err := degMinSecToDecimal(tc.inDeg, tc.inMin, tc.inSec, tc.inLet)
			if tc.expOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.expVal, v)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestConvertGPSCoordinates(t *testing.T) {
	var tcs = []struct {
		tcID      string
		inLatLong string
		expOk     bool
		expLat    float32
		expLong   float32
	}{
		{"Nominal", `1 deg 11' 60" N, 1 deg 11' 60" W`, true, 1.2, -1.2},
		{"Unparsable latitude", `b deg 11' 60" N, 1 deg 11' 60" W`, false, 0, 0},
		{"Unparsable longitude", `1 deg 11' 60" N, b deg 11' 60" W`, false, 0, 0},
		{"Wrong size", `a b`, false, 0, 0},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			lat, long, err := convertGPSCoordinates(tc.inLatLong)
			if tc.expOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.expLat, lat)
				assert.Equal(t, tc.expLong, long)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	meta := exif.FileMetadata{
		File: "aFile",
		Fields: map[string]interface{}{
			"string": "stringVal",
			"float":  float64(3.14),
			"int":    int64(42),
		},
	}

	var tcs = []struct {
		inKey     string
		expValNil bool
		expVal    string
	}{
		{"string", false, "stringVal"},
		{"float", false, "3.14"},
		{"int", false, "42"},
		{"nonExisting", true, ""},
	}

	for _, tc := range tcs {
		t.Run(tc.inKey, func(t *testing.T) {
			v := getString(meta, tc.inKey)
			if tc.expValNil {
				assert.Nil(t, v)
			} else {
				assert.NotNil(t, v)
				assert.Equal(t, tc.expVal, *v)
			}
		})
	}
}

func TestGetInt64(t *testing.T) {
	meta := exif.FileMetadata{
		File: "aFile",
		Fields: map[string]interface{}{
			"string": "stringVal",
			"float":  float64(3.14),
			"int":    int64(42),
		},
	}

	var tcs = []struct {
		inKey     string
		expValNil bool
		expVal    uint64
	}{
		{"string", true, 0},
		{"float", false, 3},
		{"int", false, 42},
		{"nonExisting", true, 0},
	}

	for _, tc := range tcs {
		t.Run(tc.inKey, func(t *testing.T) {
			v := getInt64(meta, tc.inKey)
			if tc.expValNil {
				assert.Nil(t, v)
			} else {
				assert.NotNil(t, v)
				assert.Equal(t, tc.expVal, *v)
			}
		})
	}
}

func TestGetFloat64(t *testing.T) {
	meta := exif.FileMetadata{
		File: "aFile",
		Fields: map[string]interface{}{
			"string": "stringVal",
			"float":  float64(3.14),
			"int":    int64(42),
		},
	}

	var tcs = []struct {
		inKey     string
		expValNil bool
		expVal    float64
	}{
		{"string", true, 0},
		{"float", false, 3.14},
		{"int", false, 42},
		{"nonExisting", true, 0},
	}

	for _, tc := range tcs {
		t.Run(tc.inKey, func(t *testing.T) {
			v := getFloat64(meta, tc.inKey)
			if tc.expValNil {
				assert.Nil(t, v)
			} else {
				assert.NotNil(t, v)
				assert.Equal(t, tc.expVal, *v)
			}
		})
	}
}

func TestGetStrings(t *testing.T) {
	meta := exif.FileMetadata{
		File: "aFile",
		Fields: map[string]interface{}{
			"string":  "stringVal",
			"float":   float64(3.14),
			"int":     int64(42),
			"strings": []interface{}{"str", float64(3.14), int64(42)},
		},
	}

	var tcs = []struct {
		inKey  string
		expVal []string
	}{
		{"string", []string{"stringVal"}},
		{"float", []string{"3.14"}},
		{"int", []string{"42"}},
		{"strings", []string{"str", "3.14", "42"}},
		{"nonExisting", nil},
	}

	for _, tc := range tcs {
		t.Run(tc.inKey, func(t *testing.T) {
			assert.Equal(t, tc.expVal, getStrings(meta, tc.inKey))
		})
	}
}

func TestGetDate(t *testing.T) {
	meta := exif.FileMetadata{
		File: "aFile",
		Fields: map[string]interface{}{
			"string": "stringVal",
			"date":   "2001:02:03 04:05:06",
		},
	}

	var tcs = []struct {
		inKey     string
		expValNil bool
		expVal    uint64
	}{
		{"string", false, defaultDate},
		{"date", false, 981173106000},
		{"nonExisting", false, defaultDate},
	}

	for _, tc := range tcs {
		t.Run(tc.inKey, func(t *testing.T) {
			v := getDate(meta, tc.inKey)
			if tc.expValNil {
				assert.Nil(t, v)
			} else {
				assert.NotNil(t, v)
				assert.Equal(t, tc.expVal, *v)
			}
		})
	}
}

func TestGetGPS(t *testing.T) {
	meta := exif.FileMetadata{
		File: "aFile",
		Fields: map[string]interface{}{
			"string": "stringVal",
			"gps":    `1 deg 11' 60" N, 1 deg 11' 60" W`,
		},
	}

	var tcs = []struct {
		inKey     string
		expValNil bool
		expVal    string
	}{
		{"string", true, ""},
		{"gps", false, "1.2,-1.2"},
		{"nonExisting", true, ""},
	}

	for _, tc := range tcs {
		t.Run(tc.inKey, func(t *testing.T) {
			v := getGPS(meta, tc.inKey)
			if tc.expValNil {
				assert.Nil(t, v)
			} else {
				assert.NotNil(t, v)
				assert.Equal(t, tc.expVal, *v)
			}
		})
	}
}



func TestNewMetadataExtractor(t *testing.T) {
	var tcs = []struct {
		inTC  int
		expOk bool
	}{
		{-1, false},
		{0, false},
		{2, true},
	}
	for _, tc := range tcs {
		t.Run(strconv.Itoa(tc.inTC), func(t *testing.T) {
			me, err := NewMetadataExtractor(tc.inTC)
			if tc.expOk {
				defer me.Close()
				assert.Nil(t, err)
				assert.Equal(t, tc.inTC, me.threadCount)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestNewMetadataExtractor_ErrorOnOpts(t *testing.T) {
	_, err := NewMetadataExtractor(4,func(*MetadataExtractor)error {
		return fmt.Errorf("anError")
	})
	assert.NotNil(t, err)
}

func checkTestdataPictureResult(t *testing.T, m PictureMetadata) {
	assert.Equal(t, 1.7, *m.Aperture)
	assert.Equal(t, "1/10", *m.ShutterSpeed)
	assert.Equal(t, []string{"keyword"}, m.Keywords)
	assert.Equal(t, "model", *m.CameraModel)
	assert.Equal(t, "lensmodel", *m.LensModel)
	assert.Equal(t, "image/jpeg", *m.MimeType)
	assert.Equal(t, uint64(550), *m.Height)
	assert.Equal(t, uint64(458), *m.Width)
	assert.Equal(t, uint64(20504), m.FileSize)
	assert.Equal(t, uint64(1571912945000), *m.Date)
	assert.Equal(t, "picture.jpg", m.FileName)
	assert.Equal(t, "testdata", m.Folder)
}

func TestExtractMetadataFromFileNominal(t *testing.T) {
	ext, err := NewMetadataExtractor(4)
	assert.Nil(t, err)
	defer ext.Close()

	f := "../../testdata/picture.jpg"
	fInfo, err := os.Stat(f)
	assert.Nil(t, err)
	m, err := ext.extractMetadataFromFile(context.TODO(), f, fInfo)

	assert.Nil(t, err)
	checkTestdataPictureResult(t, m)
}

func TestExtractMetadata(t *testing.T) {
	inChan := make(chan browse.Task, 10)
	outChan := make(chan PictureMetadata, 10)

	f := "../../testdata/picture.jpg"
	fInfo, err := os.Stat(f)
	assert.Nil(t, err)
	inChan <- browse.Task{Path: f, Info: fInfo}
	close(inChan)

	ext, err := NewMetadataExtractor(4)
	assert.Nil(t, err)
	err = ext.ExtractMetadata(context.TODO(), inChan, outChan)
	assert.Nil(t, err)

	pics := []PictureMetadata{}
	for cur := range outChan {
		pics = append(pics, cur)
	}
	assert.Equal(t, 1, len(pics))
	checkTestdataPictureResult(t, pics[0])
}
