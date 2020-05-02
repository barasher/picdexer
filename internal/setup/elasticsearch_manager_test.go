package setup

import (
	"github.com/barasher/picdexer/conf"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPutMapping_Nominal(t *testing.T) {
	expBody, err := readFile(t, "./assets/mapping.json")
	assert.Nil(t, err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/picdexer", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, expBody, string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c :=
		conf.ElasticsearchConf{
			Url: ts.URL,
		}

	s, err := NewESManager(c)
	assert.Nil(t, err)
	assert.Nil(t, s.PutMapping(getHttpClient()))
}

func TestPutMapping_WrongStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	c := conf.ElasticsearchConf{
		Url: ts.URL,
	}
	s, err := NewESManager(c)
	assert.Nil(t, err)
	assert.NotNil(t, s.PutMapping(getHttpClient()))
}

func TestPutMapping_FailRequest(t *testing.T) {
	c := conf.ElasticsearchConf{
		Url: "blablabla",
	}
	s, err := NewESManager(c)
	assert.Nil(t, err)
	assert.NotNil(t, s.PutMapping(getHttpClient()))
}

func TestSimpleMappingQuery(t *testing.T) {
	var tcs = []struct {
		tcID          string
		tcInExpStatus int
		tcExpOk       bool
	}{
		{"ok", http.StatusNoContent, true},
		{"ko", http.StatusOK, false},
	}
	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/picdexer", r.URL.Path)
				w.WriteHeader(http.StatusNoContent)
			}))
			defer ts.Close()
			c := conf.ElasticsearchConf{Url: ts.URL}
			h := http.Client{Timeout: 10 * time.Second}
			s, err := NewESManager(c)
			assert.Nil(t, err)
			err = s.simpleMappingQuery(&h, http.MethodGet, tc.tcInExpStatus)
			assert.Equal(t, tc.tcExpOk, err == nil)
			t.Logf("status: %v", err)
		})
	}
}