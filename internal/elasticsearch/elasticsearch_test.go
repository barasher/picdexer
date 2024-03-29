package elasticsearch

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/internal/metadata"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
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
		b, err := io.ReadAll(r.Body)
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
	return EsDoc{
		Header: EsHeader{
			Index: EsHeaderIndex{
				Index: "idx",
				ID:    id,
			},
		},
		Document: metadata.PictureMetadata{
			FileName: filename,
		},
	}
}

func TestPush_Nominal(t *testing.T) {
	collectedBodies := []string{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/_bulk", r.URL.Path)
		assert.Equal(t, "application/x-ndjson", r.Header.Get("Content-type"))
		b, err := io.ReadAll(r.Body)
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
	assert.Equal(t, "{\"index\":{\"_index\":\"idx\",\"_id\":\"id1\"}}\n"+
		"{\"FileName\":\"f1.jpg\",\"Folder\":\"\",\"ImportID\":\"\",\"FileSize\":0}\n"+
		"{\"index\":{\"_index\":\"idx\",\"_id\":\"id2\"}}\n"+
		"{\"FileName\":\"f2.jpg\",\"Folder\":\"\",\"ImportID\":\"\",\"FileSize\":0}\n", collectedBodies[0])
	assert.Equal(t, "{\"index\":{\"_index\":\"idx\",\"_id\":\"id3\"}}\n"+
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
	// {"index":{"_index":"idx","_id":"id1"}}
	// {"FileName":"f1.jpg","Folder":"","ImportID":"","FileSize":0}
	// {"index":{"_index":"idx","_id":"id2"}}
	// {"FileName":"f2.jpg","Folder":"","ImportID":"","FileSize":0}
	// {"index":{"_index":"idx","_id":"id3"}}
	// {"FileName":"f3.jpg","Folder":"","ImportID":"","FileSize":0}
}

func TestConvertMetadataToEsDoc(t *testing.T) {
	in := make(chan metadata.PictureMetadata, 2)
	in <- metadata.PictureMetadata{
		FileName:   "picture.jpg",
		SourceFile: "../../testdata/picture.jpg",
		FileID:     "fileIDValue",
	}
	close(in)

	out := make(chan EsDoc, 2)
	p, _ := NewEsPusher(10)
	p.ConvertMetadataToEsDoc(context.TODO(), in, out)

	docs := []EsDoc{}
	for cur := range out {
		docs = append(docs, cur)
	}

	assert.Equal(t, 1, len(docs))
	doc, ok := docs[0].Document.(metadata.PictureMetadata)
	assert.True(t, ok)
	assert.Equal(t, "../../testdata/picture.jpg", doc.SourceFile)
	assert.Equal(t, "picture.jpg", doc.FileName)
	assert.Equal(t, "fileIDValue", docs[0].Header.Index.ID)
	assert.Equal(t, "picdexer", docs[0].Header.Index.Index)
}

func TestConvertMetadataToEsDoc_WithSync(t *testing.T) {
	in := make(chan metadata.PictureMetadata, 2)
	d, err := time.Parse("2006:01:02", "2021:01:01")
	ts := uint64(d.Unix())
	assert.Nil(t,err)
	in <- metadata.PictureMetadata{
		FileName:   "f1.jpg",
		SourceFile: "../../testdata/picture.jpg",
		FileID:     "f1IDValue",
		Keywords: []string{"kw1", "kw2"},
		Date: &ts,
	}
	close(in)

	out := make(chan EsDoc, 5)
	sync, err := time.Parse("2006:01:02", "2020/01/01")
	p, _ := NewEsPusher(10, SyncOnDate("kw2", sync))
	p.ConvertMetadataToEsDoc(context.TODO(), in, out)

	docs := []EsDoc{}
	for cur := range out {
		docs = append(docs, cur)
	}

	assert.Equal(t, 2, len(docs))
	doc, ok := docs[1].Document.(SyncOnDateBody)
	assert.True(t, ok)
	assert.Equal(t, uint64(63083891059200), uint64(doc.SyncedDate))
	assert.Equal(t, uint64(d.Unix()), doc.Date)
	assert.Equal(t, "kw2_f1IDValue", docs[1].Header.Index.ID)
	assert.Equal(t, "sync-on-date", docs[1].Header.Index.Index)
}