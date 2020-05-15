package metadata

import (
	"bytes"
	"context"
	"github.com/barasher/picdexer/conf"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertNominal(t *testing.T) {
	idxer, err := NewIndexer(conf.ElasticsearchConf{})
	assert.Nil(t, err)
	defer idxer.Close()

	f := "../../testdata/picture.jpg"
	fInfo, err := os.Stat(f)
	assert.Nil(t, err)
	m, err := idxer.convert(context.Background(), f, fInfo)

	assert.Nil(t, err)
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



func TestDumpNominal(t *testing.T) {
	p, err := NewIndexer(conf.ElasticsearchConf{})
	err = p.Dump(context.Background(), "../../testdata", os.Stdout)
	assert.Nil(t, err)
}

func TestPushNominal(t *testing.T) {
	content := "content"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, ndJsonMimeType, req.Header.Get("Content-Type"))
		body, _ := ioutil.ReadAll(req.Body)
		assert.Equal(t, content, string(body))
		rw.WriteHeader(200)
	}))
	defer server.Close()

	c := conf.ElasticsearchConf{Url: server.URL}
	idxer, err := NewIndexer(c)
	assert.Nil(t, err)
	defer idxer.Close()

	buf := bytes.NewBufferString(content)
	err = idxer.Push(context.Background(), buf)
	assert.Nil(t, err)
}

func TestPushWrongStatusCode(t *testing.T) {
	content := "content"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(500)
	}))
	defer server.Close()

	c := conf.ElasticsearchConf{Url: server.URL}
	idxer, err := NewIndexer(c)
	assert.Nil(t, err)
	defer idxer.Close()

	buf := bytes.NewBufferString(content)
	err = idxer.Push(context.Background(), buf)
	assert.NotNil(t, err)
}

func TestPushPostFailure(t *testing.T) {
	idxer, err := NewIndexer(conf.ElasticsearchConf{})
	assert.Nil(t, err)
	defer idxer.Close()

	buf := bytes.NewBufferString("blabla")
	err = idxer.Push(context.Background(), buf)
	assert.NotNil(t, err)
}

func TestExtractionThreadCount(t *testing.T) {
	var tcs = []struct {
		tcID     string
		inConfValue      int
		expValue int
	}{
		{"-1", -1, defaultExtrationThreadCount},
		{"0", 0, defaultExtrationThreadCount},
		{"5", 5,5},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			c := conf.ElasticsearchConf{ExtractionThreadCount:tc.inConfValue}
			i := Indexer{conf: c}
			assert.Equal(t, tc.expValue, i.extractionThreadCount())
		})
	}
}

func TestToExtractChannelSize(t *testing.T) {
	var tcs = []struct {
		tcID     string
		inConfValue      int
		expValue int
	}{
		{"-1", -1, defaultToExtractChannelSize},
		{"0", 0, defaultToExtractChannelSize},
		{"5", 5,5},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			c := conf.ElasticsearchConf{ToExtractChannelSize:tc.inConfValue}
			i := Indexer{conf: c}
			assert.Equal(t, tc.expValue, i.toExtractChannelSize())
		})
	}
}