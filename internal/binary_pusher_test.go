package internal

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNopPusher(t *testing.T) {
	p := NewNopPusher()
	assert.Nil(t, p.push("k", "v"))
}

func TestPusher_StatusCode(t *testing.T) {
	var tcs = []struct {
		tcID       string
		httpCode   int
		expSuccess bool
	}{
		{"Nominal", http.StatusNoContent, true},
		{"Wrong", http.StatusInternalServerError, false},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.httpCode)
			}))
			defer ts.Close()

			err := NewPusher(ts.URL).push("../testdata/picture.jpg", "myKey")
			assert.Equal(t, tc.expSuccess, err == nil)
		})
	}
}

func TestPusher_UnknownFile(t *testing.T) {
	err := NewPusher("").push("../testdata/unknown.jpg", "myKey")
	t.Logf("err: %v", err)
	assert.NotNil(t, err)
}


func TestPusher_WrongUrl(t *testing.T) {
	err := NewPusher("file:/tmp/").push("../testdata/picture.jpg", "myKey")
	t.Logf("err: %v", err)
	assert.NotNil(t, err)
}
