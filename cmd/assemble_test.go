package cmd

import (
	"context"
	"github.com/barasher/picdexer/internal/binary"
	"github.com/stretchr/testify/assert"
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

	resizeDir, err := os.MkdirTemp(os.TempDir(), "picdexer")
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
			Url: esServer.URL,
		},
		Binary: BinaryConf{
			Url:        binServer.URL,
			Height:     50,
			Width:      70,
			WorkingDir: resizeDir,
		},
	}

	err = Run(context.Background(), c, []string{"../testdata/"})
	assert.Nil(t, err)
	assert.True(t, esDocPushed)
	assert.True(t, binPushed)
}

func TestMax(t *testing.T) {
	assert.Equal(t, 2, max(1, 2))
	assert.Equal(t, 2, max(2, 1))
	assert.Equal(t, 2, max(2, 2))
}

func TestBuildBinaryManager_Lazy(t *testing.T) {
	bm, _, err := buildBinaryManager(
		Config{
			Binary: BinaryConf{
				Url: "",
			},
		})
	assert.Nil(t, err)
	assert.IsType(t, binary.LazyBinaryManager{}, bm)
}
