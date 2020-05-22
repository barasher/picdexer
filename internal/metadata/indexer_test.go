package metadata

import (
	"bytes"
	"context"
	"github.com/barasher/picdexer/conf"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const sampleTaskChanResult = `{"index":{"_index":"idx","_id":"id"}}
{"FileName":"toto.jpg","Folder":"","ImportID":"","FileSize":0}
`

const nominalOutput = `{"index":{"_index":"picdexer","_id":"ec3d25618be7af41c6824855f0f42c73_picture.jpg"}}
{"FileName":"picture.jpg","Folder":"testdata","ImportID":"","FileSize":20504,"Aperture":1.7,"ShutterSpeed":"1/10","Keywords":["keyword"],"CameraModel":"model","LensModel":"lensmodel","MimeType":"image/jpeg","Height":550,"Width":458,"Date":1571912945000}
`

func TestConvertNominal(t *testing.T) {
	idxer, err := NewIndexer(conf.ElasticsearchConf{})
	assert.Nil(t, err)
	defer idxer.Close()

	f := "../../testdata/picture.jpg"
	fInfo, err := os.Stat(f)
	assert.Nil(t, err)
	m, err := idxer.convert(context.Background(), f, fInfo)

	assert.Nil(t, err)
	assert.Equal(t, 1.7, *m.Aperture)
	assert.Equal(t, "1/10", *m.ShutterSpeed)
	assert.Equal(t, []string{"keyword"}, m.Keywords)
	assert.Equal(t, "model", *m.CameraModel)
	assert.Equal(t, "lensmodel", *m.LensModel)
	assert.Equal(t, "image/jpeg", *m.MimeType)
	assert.Equal(t, uint64(550), *m.Height)
	assert.Equal(t, uint64(458), *m.Width)
	assert.Equal(t, uint64(20504), m.FileSize)
	assert.Equal(t, uint64(1571912945000), *m.Date)
	assert.Equal(t, "picture.jpg", m.FileName)
	assert.Equal(t, "testdata", m.Folder)
}

func TestExtractionThreadCount(t *testing.T) {
	var tcs = []struct {
		tcID        string
		inConfValue int
		expValue    int
	}{
		{"-1", -1, defaultExtrationThreadCount},
		{"0", 0, defaultExtrationThreadCount},
		{"5", 5, 5},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			c := conf.ElasticsearchConf{ExtractionThreadCount: tc.inConfValue}
			i := Indexer{conf: c}
			assert.Equal(t, tc.expValue, i.extractionThreadCount())
		})
	}
}

func TestToExtractChannelSize(t *testing.T) {
	var tcs = []struct {
		tcID        string
		inConfValue int
		expValue    int
	}{
		{"-1", -1, defaultToExtractChannelSize},
		{"0", 0, defaultToExtractChannelSize},
		{"5", 5, 5},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			c := conf.ElasticsearchConf{ToExtractChannelSize: tc.inConfValue}
			i := Indexer{conf: c}
			assert.Equal(t, tc.expValue, i.toExtractChannelSize())
		})
	}
}

func buildCollectTaskChan() chan printTask {
	h := bulkEntryHeader{}
	h.Index.Index = "idx"
	h.Index.ID = "id"
	task := printTask{
		header: h,
		pic:    Model{FileName: "toto.jpg"},
	}
	taskChan := make(chan printTask, 1)
	taskChan <- task
	return taskChan
}

func buildConvertTaskChan(t *testing.T) chan ExtractTask {
	f := "../../testdata/picture.jpg"
	fInfo, err := os.Stat(f)
	assert.Nil(t, err)
	task := ExtractTask{
		Path: f,
		Info: fInfo,
	}
	taskChan := make(chan ExtractTask, 1)
	taskChan <- task
	return taskChan
}

func TestCollectToPrint(t *testing.T) {
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	taskChan := buildCollectTaskChan()
	close(taskChan)

	idxer := Indexer{}
	ctx, cancel := context.WithCancel(context.Background())
	err := idxer.collectToPrint(ctx, cancel, taskChan)
	assert.Nil(t, err)
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	assert.Equal(t, sampleTaskChanResult, buf.String())
}

func TestCollectToPushNominal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/_bulk", r.URL.Path)
		assert.Equal(t, "application/x-ndjson", r.Header.Get("Content-type"))
		b, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		defer r.Body.Close()
		assert.Equal(t, sampleTaskChanResult, string(b))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	taskChan := buildCollectTaskChan()
	close(taskChan)

	idxer, err := NewIndexer(conf.ElasticsearchConf{Url: ts.URL})
	assert.Nil(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	err = idxer.collectToPush(ctx, cancel, taskChan)
	assert.Nil(t, err)
}

func TestCollectToPushBadHttpStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	taskChan := buildCollectTaskChan()
	close(taskChan)

	idxer, err := NewIndexer(conf.ElasticsearchConf{Url: ts.URL})
	assert.Nil(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	err = idxer.collectToPush(ctx, cancel, taskChan)
	assert.NotNil(t, err)
}

func TestExtractFolder(t *testing.T) {
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	idxer, err := NewIndexer(conf.ElasticsearchConf{})
	assert.Nil(t, err)
	err = idxer.ExtractFolder(context.Background(), "../../testdata")
	assert.Nil(t, err)

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	assert.Equal(t, nominalOutput, buf.String())
}

func TestExtractAndPushFolder(t *testing.T) {
	pushed := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushed = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	idxer, err := NewIndexer(conf.ElasticsearchConf{Url: ts.URL})
	assert.Nil(t, err)
	err = idxer.ExtractAndPushFolder(context.Background(), "../../testdata")
	assert.Nil(t, err)
	assert.True(t, pushed)
}

func TestExtractTasks(t *testing.T) {
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	taskChan := buildConvertTaskChan(t)
	close(taskChan)

	p, err := NewIndexer(conf.ElasticsearchConf{})
	assert.Nil(t, err)
	err = p.ExtractTasks(context.Background(), taskChan)
	assert.Nil(t, err)

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	assert.Equal(t, nominalOutput, buf.String())
}

func TestExtractAndPushTasks(t  *testing.T) {
	pushed := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushed = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()


	taskChan := buildConvertTaskChan(t)
	close(taskChan)

	idxer, err := NewIndexer(conf.ElasticsearchConf{Url: ts.URL})
	assert.Nil(t, err)
	err = idxer.ExtractAndPushTasks(context.Background(), taskChan)
	assert.Nil(t, err)
	assert.True(t, pushed)
}

