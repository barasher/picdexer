package internal

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBulkSize(t *testing.T) {
	var tcs = []struct {
		tcID        string
		inConfValue int
		expValue    int
	}{
		{"-1", -1, defaultBulkSize},
		{"0", 0, defaultBulkSize},
		{"5", 5, 5},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			c := EsPusherConf{BulkSize: tc.inConfValue}
			p, err := NewEsPusher(c)
			assert.Nil(t, err)
			assert.Equal(t, tc.expValue, p.bulkSize())
		})
	}
}

func TestPushToEs_Nominal(t *testing.T) {
	expBody := "sampleBody"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/_bulk", r.URL.Path)
		assert.Equal(t, "application/x-ndjson", r.Header.Get("Content-type"))
		b, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, expBody, string(b))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	pusher, err := NewEsPusher(EsPusherConf{Url: ts.URL})
	assert.Nil(t, err)
	assert.Nil(t, pusher.pushToEs(context.TODO(), strings.NewReader(expBody)))
}

func TestPushToEs_BadHttpStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	pusher, err := NewEsPusher(EsPusherConf{Url: ts.URL})
	assert.Nil(t, err)
	assert.NotNil(t, pusher.pushToEs(context.TODO(), strings.NewReader("bla")))
}

func buildEsDoc(id string, filename string) EsDoc {
	d := EsDoc{}
	d.Header.Index = "idx"
	d.Header.ID = id
	d.Document = PictureMetadata{FileName: filename}
	return d
}

func TestPush_Nominal(t *testing.T) {
	collectedBodies := []string{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/_bulk", r.URL.Path)
		assert.Equal(t, "application/x-ndjson", r.Header.Get("Content-type"))
		b, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		collectedBodies = append(collectedBodies, string(b))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	pusher, err := NewEsPusher(EsPusherConf{Url: ts.URL, BulkSize: 2})
	assert.Nil(t, err)

	inChan := make(chan EsDoc, 4)
	inChan <- buildEsDoc("id1", "f1.jpg")
	inChan <- buildEsDoc("id2", "f2.jpg")
	inChan <- buildEsDoc("id3", "f3.jpg")
	close(inChan)

	assert.Nil(t, pusher.Push(context.TODO(), inChan))
	assert.Equal(t, 2, len(collectedBodies))
	assert.Equal(t, "{\"_index\":\"idx\",\"_id\":\"id1\"}\n"+
		"{\"FileName\":\"f1.jpg\",\"Folder\":\"\",\"ImportID\":\"\",\"FileSize\":0}\n"+
		"{\"_index\":\"idx\",\"_id\":\"id2\"}\n"+
		"{\"FileName\":\"f2.jpg\",\"Folder\":\"\",\"ImportID\":\"\",\"FileSize\":0}\n", collectedBodies[0])
	assert.Equal(t, "{\"_index\":\"idx\",\"_id\":\"id3\"}\n"+
		"{\"FileName\":\"f3.jpg\",\"Folder\":\"\",\"ImportID\":\"\",\"FileSize\":0}\n", collectedBodies[1])
}

func TestPush_ErrorOnPush(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	pusher, err := NewEsPusher(EsPusherConf{Url: ts.URL, BulkSize: 2})
	assert.Nil(t, err)

	inChan := make(chan EsDoc, 1)
	inChan <- buildEsDoc("id1", "f1.jpg")
	close(inChan)

	assert.NotNil(t, pusher.Push(context.TODO(), inChan))
}

func TestPush_ErrorOnPush2(t *testing.T) {
	q := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if q == 1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		q++
	}))
	defer ts.Close()

	pusher, err := NewEsPusher(EsPusherConf{Url: ts.URL, BulkSize: 2})
	assert.Nil(t, err)

	inChan := make(chan EsDoc, 3)
	inChan <- buildEsDoc("id1", "f1.jpg")
	inChan <- buildEsDoc("id2", "f2.jpg")
	inChan <- buildEsDoc("id3", "f3.jpg")
	close(inChan)

	assert.NotNil(t, pusher.Push(context.TODO(), inChan))
}
