package setup

import (
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
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

func getHttpClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

type esManagerMock struct {
	maeBool   bool
	maeErr    error
	dmErr     error
	pmErr     error
	maeCalled bool
	dmCalled  bool
	pmCalled  bool
}

func (e *esManagerMock) MappingAlreadyExist(client *http.Client) (bool, error) {
	e.maeCalled = true
	return e.maeBool, e.maeErr
}

func (e *esManagerMock) DeleteMapping(client *http.Client) error {
	e.dmCalled = true
	return e.dmErr
}

func (e *esManagerMock) PutMapping(client *http.Client) error {
	e.pmCalled = true
	return e.pmErr
}

func NewESMMock() *esManagerMock {
	return &esManagerMock{}
}

func (e *esManagerMock) setMAE(b bool, err error) *esManagerMock {
	e.maeBool, e.maeErr = b, err
	return e
}

func (e *esManagerMock) setDM(err error) *esManagerMock {
	e.dmErr = err
	return e
}

func (e *esManagerMock) setPM(err error) *esManagerMock {
	e.pmErr = err
	return e
}

func (e *esManagerMock) checkCalled(t *testing.T, expMAE bool, expDM bool, expPM bool) {
	assert.Equal(t, expMAE, e.maeCalled)
	assert.Equal(t, expDM, e.dmCalled)
	assert.Equal(t, expPM, e.pmCalled)
}

func TestSetupElasticsearch_OkWithoutMapping(t *testing.T) {
	esm := NewESMMock().setMAE(false, nil)
	s := &Setup{}
	assert.Nil(t, s.setupElasticsearch(esm))
	esm.checkCalled(t, true, false, true)
}

func TestSetupElasticsearch_OkWithMapping(t *testing.T) {
	esm := NewESMMock().setMAE(true, nil)
	s := &Setup{}
	assert.Nil(t, s.setupElasticsearch(esm))
	esm.checkCalled(t, true, true, true)
}

func TestSetupElasticsearch_FailOnExistenceCheck(t *testing.T) {
	esm := NewESMMock().setMAE(false, fmt.Errorf("e"))
	s := &Setup{}
	assert.NotNil(t, s.setupElasticsearch(esm))
	esm.checkCalled(t, true, false, false)
}

func TestSetupElasticsearch_FailOnDeleteMapping(t *testing.T) {
	esm := NewESMMock().setMAE(true, nil).setDM(fmt.Errorf("e"))
	s := &Setup{}
	assert.NotNil(t, s.setupElasticsearch(esm))
	esm.checkCalled(t, true, true, false)
}

func TestSetupElasticsearch_FailOnPutMapping(t *testing.T) {
	esm := NewESMMock().setMAE(false, nil).setPM(fmt.Errorf("e"))
	s := &Setup{}
	assert.NotNil(t, s.setupElasticsearch(esm))
	esm.checkCalled(t, true, false, true)
}

func TestSetupKibana_Nominal(t *testing.T) {
	expBody, err := readFile(t, "./assets/kibana.ndjson")
	assert.Nil(t, err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/saved_objects/_import", r.URL.Path)
		assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))
		assert.Equal(t, "true", r.Header.Get("kbn-xsrf"))
		assert.Equal(t, "true", r.URL.Query().Get("overwrite"))
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
