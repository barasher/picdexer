package elasticsearch

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/internal/metadata"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestNewEsPusher(t *testing.T) {
	var tcs = []struct {
		inBS  int
		expOk bool
	}{
		{-1, false},
		{0, false},
		{2, true},
	}
	for _, tc := range tcs {
		t.Run(strconv.Itoa(tc.inBS), func(t *testing.T) {
			p, err := NewEsPusher(tc.inBS)
			if tc.expOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.inBS, p.bulkSize)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestNewEsPusher_ErrorOnOpts(t *testing.T) {
	_, err := NewEsPusher(50, func(*EsPusher) error {
		return fmt.Errorf("anError")
	})
	assert.NotNil(t, err)
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

	pusher, err := NewEsPusher(50, EsUrl(ts.URL))
	assert.Nil(t, err)
	assert.Nil(t, pusher.pushToEs(context.TODO(), strings.NewReader(expBody)))
}

func TestPushToEs_BadHttpStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	pusher, err := NewEsPusher(50, EsUrl(ts.URL))
	assert.Nil(t, err)
	assert.NotNil(t, pusher.pushToEs(context.TODO(), strings.NewReader("bla")))
}

func buildEsDoc(id string, filename string) EsDoc {
	d := EsDoc{}
	d.Header.Index = "idx"
	d.Header.ID = id
	d.Document = metadata.PictureMetadata{FileName: filename}
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

	pusher, err := NewEsPusher(2, EsUrl(ts.URL))
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

	pusher, err := NewEsPusher(2, EsUrl(ts.URL))
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

	pusher, err := NewEsPusher(2, EsUrl(ts.URL))
	assert.Nil(t, err)

	inChan := make(chan EsDoc, 3)
	inChan <- buildEsDoc("id1", "f1.jpg")
	inChan <- buildEsDoc("id2", "f2.jpg")
	inChan <- buildEsDoc("id3", "f3.jpg")
	close(inChan)

	assert.NotNil(t, pusher.Push(context.TODO(), inChan))
}

func ExamplePrint() {
	pusher, err := NewEsPusher(2, EsUrl("a"))
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	inChan := make(chan EsDoc, 4)
	inChan <- buildEsDoc("id1", "f1.jpg")
	inChan <- buildEsDoc("id2", "f2.jpg")
	inChan <- buildEsDoc("id3", "f3.jpg")
	close(inChan)

	err = pusher.Print(context.TODO(), inChan)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	// Output:
	// {"_index":"idx","_id":"id1"}
	// {"FileName":"f1.jpg","Folder":"","ImportID":"","FileSize":0}
	// {"_index":"idx","_id":"id2"}
	// {"FileName":"f2.jpg","Folder":"","ImportID":"","FileSize":0}
	// {"_index":"idx","_id":"id3"}
	// {"FileName":"f3.jpg","Folder":"","ImportID":"","FileSize":0}
}

func TestConvertMetadataToEsDoc(t *testing.T) {
	in := make(chan metadata.PictureMetadata, 2)
	in <- metadata.PictureMetadata{
		FileName:   "picture.jpg",
		SourceFile: "../../testdata/picture.jpg",
	}
	in <- metadata.PictureMetadata{
		FileName:   "nonExisting.jpg",
		SourceFile: "../testdata/nonExisting.jpg",
	}
	close(in)

	out := make(chan EsDoc, 2)
	ConvertMetadataToEsDoc(context.TODO(), in, out)

	docs := []EsDoc{}
	for cur := range out {
		docs = append(docs, cur)
	}

	assert.Equal(t, 1, len(docs))
	doc, ok := docs[0].Document.(metadata.PictureMetadata)
	assert.True(t, ok)
	assert.Equal(t, "../../testdata/picture.jpg", doc.SourceFile)
	assert.Equal(t, "picture.jpg", doc.FileName)
	assert.Equal(t, "ec3d25618be7af41c6824855f0f42c73_picture.jpg", docs[0].Header.ID)
	assert.Equal(t, "picdexer", docs[0].Header.Index)
}

func TestGetID(t *testing.T) {
	var tcs = []struct {
		tcID  string
		inF   string
		expOK bool
		expID string
	}{
		{"nominal", "../../testdata/picture.jpg", true, "ec3d25618be7af41c6824855f0f42c73_picture.jpg"},
		{"nonExisting", "nonExisting", false, ""},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			i, err := getID(tc.inF)
			if tc.expOK {
				assert.Nil(t, err)
				assert.Equal(t, tc.expID, i)
			} else {
				assert.NotNil(t, err)
			}
		})
	}

}