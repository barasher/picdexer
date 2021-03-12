package setup

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"io"
)

func TestPutMapping_Nominal(t *testing.T) {
	expBody, err := readFile(t, "./assets/mapping.json")
	assert.Nil(t, err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/picdexer", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, expBody, string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s, err := NewESManager(ts.URL)
	assert.Nil(t, err)
	assert.Nil(t, s.PutMapping(getHttpClient()))
}

func TestPutMapping_WrongStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	s, err := NewESManager(ts.URL)
	assert.Nil(t, err)
	assert.NotNil(t, s.PutMapping(getHttpClient()))
}

func TestPutMapping_FailRequest(t *testing.T) {
	s, err := NewESManager("blablabla")
	assert.Nil(t, err)
	assert.NotNil(t, s.PutMapping(getHttpClient()))
}

func TestSimpleMappingQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/picdexer", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	h := http.Client{Timeout: 10 * time.Second}
	s, err := NewESManager(ts.URL)
	assert.Nil(t, err)
	status, _ := s.simpleMappingQuery(&h, http.MethodGet)
	assert.Equal(t, http.StatusNoContent, status)
}

func TestMappingAlreadyExist(t *testing.T) {
	var tcs = []struct {
		tcID             string
		inReturnedStatus int
		expSuccess       bool
		expExists        bool
	}{
		{"200", http.StatusOK, true, true},
		{"404", http.StatusNotFound, true, false},
		{"500", http.StatusInternalServerError, false, false},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.inReturnedStatus)
			}))
			defer ts.Close()
			h := http.Client{Timeout: 10 * time.Second}
			s, err := NewESManager(ts.URL)
			assert.Nil(t, err)
			exists, err := s.MappingAlreadyExist(&h)
			if tc.expSuccess {
				assert.Nil(t, err)
				assert.Equal(t, tc.expExists, exists)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestDeleteMapping(t *testing.T) {
	var tcs = []struct {
		tcID             string
		inReturnedStatus int
		expSuccess       bool
	}{
		{"200", http.StatusOK, true},
		{"500", http.StatusInternalServerError, false},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.inReturnedStatus)
			}))
			defer ts.Close()
			h := http.Client{Timeout: 10 * time.Second}
			s, err := NewESManager(ts.URL)
			assert.Nil(t, err)
			err = s.DeleteMapping(&h)
			if tc.expSuccess {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
