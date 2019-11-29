package indexer

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertNominal(t *testing.T) {
	idxer, err := NewIndexer()
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

func TestNewIndexerNominal(t *testing.T) {
	v1 := false
	f1 := func(idxer *Indexer) error {
		v1 = true
		return nil
	}
	v2 := false
	f2 := func(idxer *Indexer) error {
		v2 = true
		return nil
	}
	idxer, err := NewIndexer(f1, f2)
	assert.Nil(t, err)
	defer idxer.Close()
	assert.True(t, v1)
	assert.True(t, v2)
}

func TestNewIndexerFailureOnOption(t *testing.T) {
	f := func(idxer *Indexer) error {
		return fmt.Errorf("error")
	}
	_, err := NewIndexer(f)
	assert.NotNil(t, err)
}

func TestInput(t *testing.T) {
	idxer, err := NewIndexer(Input("toto"))
	assert.Nil(t, err)
	assert.Equal(t, "toto", idxer.input)
}

func TestDumpNominal(t *testing.T) {
	p, err := NewIndexer(Input("../../testdata"))
	err = p.Dump(context.Background(), os.Stdout)
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

	idxer, err := NewIndexer()
	assert.Nil(t, err)
	defer idxer.Close()

	buf := bytes.NewBufferString(content)
	err = idxer.Push(context.Background(), server.URL, buf)
	assert.Nil(t, err)
}

func TestPushWrongStatusCode(t *testing.T) {
	content := "content"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(500)
	}))
	defer server.Close()

	idxer, err := NewIndexer()
	assert.Nil(t, err)
	defer idxer.Close()

	buf := bytes.NewBufferString(content)
	err = idxer.Push(context.Background(), server.URL, buf)
	assert.NotNil(t, err)
}

func TestPushPostFailure(t *testing.T) {
	idxer, err := NewIndexer()
	assert.Nil(t, err)
	defer idxer.Close()

	buf := bytes.NewBufferString("blabla")
	err = idxer.Push(context.Background(), "aaa", buf)
	assert.NotNil(t, err)
}
