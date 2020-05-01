package setup

import (
	"github.com/barasher/picdexer/conf"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func readFile(t *testing.T, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func TestSetupElasticsearch_Nominal(t *testing.T) {
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

	c := conf.Conf{
		Elasticsearch: conf.ElasticsearchConf{
			Url: ts.URL,
		},
	}
	s, err := NewSetup(c)
	assert.Nil(t, err)
	assert.Nil(t, s.SetupElasticsearch())
}

func TestSetupElasticsearch_WrongStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	c := conf.Conf{
		Elasticsearch: conf.ElasticsearchConf{
			Url: ts.URL,
		},
	}
	s, err := NewSetup(c)
	assert.Nil(t, err)
	assert.NotNil(t, s.SetupElasticsearch())
}

func TestSetupElasticsearch_FailRequest(t *testing.T) {
	c := conf.Conf{
		Elasticsearch: conf.ElasticsearchConf{
			Url: "blablabla",
		},
	}
	s, err := NewSetup(c)
	assert.Nil(t, err)
	assert.NotNil(t, s.SetupElasticsearch())
}

func TestSetupKibana_Nominal(t *testing.T) {
	expBody, err := readFile(t, "./assets/kibana.ndjson")
	assert.Nil(t, err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/saved_objects/_import", r.URL.Path)
		assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))
		assert.Equal(t, "true", r.Header.Get("kbn-xsrf"))
		f, _, err := r.FormFile("file")
		t.Logf("formFile error: %s", err)
		assert.Nil(t, err)
		if err == nil {
			defer f.Close()
		}
		b, err := ioutil.ReadAll(f)
		assert.Nil(t, err)
		assert.Equal(t, expBody, string(b))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := conf.Conf{
		Kibana: conf.KibanaConf{
			Url: ts.URL,
		},
	}
	s, err := NewSetup(c)
	assert.Nil(t, err)
	assert.Nil(t, s.SetupKibana())
}

func TestSetupKibana_WrongStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	c := conf.Conf{
		Kibana: conf.KibanaConf{
			Url: ts.URL,
		},
	}
	s, err := NewSetup(c)
	assert.Nil(t, err)
	assert.NotNil(t, s.SetupKibana())
}

func TestSetupKibana_FailRequest(t *testing.T) {
	c := conf.Conf{
		Kibana: conf.KibanaConf{
			Url: "blablabla",
		},
	}
	s, err := NewSetup(c)
	assert.Nil(t, err)
	assert.NotNil(t, s.SetupKibana())
}
