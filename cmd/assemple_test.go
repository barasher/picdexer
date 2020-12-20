package cmd

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	binPushed := false
	binServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		binPushed = true
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Logf("binServer url: %s", binServer.URL)
	defer binServer.Close()

	resizeDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	assert.Nil(t, err)
	t.Logf("temp folder: %s", resizeDir)
	defer os.RemoveAll(resizeDir)

	esDocPushed := false
	esServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		esDocPushed = true
		w.WriteHeader(http.StatusOK)
	}))
	defer esServer.Close()
	t.Logf("esServer url: %s", esServer.URL)

	c := Config{
		Elasticsearch: ElasticsearchConf{
			Url:         esServer.URL,
		},
		Binary:        BinaryConf{
			Url:         binServer.URL,
			Height:      50,
			Width:       70,
			WorkingDir:  resizeDir,
		},
	}

	err = Run(c, "../testdata/")
	assert.Nil(t, err)
	assert.True(t, esDocPushed)
	assert.True(t, binPushed)
}
