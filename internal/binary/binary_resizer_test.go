package binary

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNopResizerResize(t *testing.T) {
	r := NewNopResizer()
	assert.Nil(t, r.resize(context.TODO(), "../../testdata/picture.jpg", "/tmp/a.jpg"))
}

func TestNopResizerCleanUp(t *testing.T) {
	assert.Nil(t, NewNopResizer().cleanup(context.TODO(), "blabla"))
}

func TestResizer_Nominal(t *testing.T) {
	outDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	assert.Nil(t, err)
	t.Logf("temp folder: %s", outDir)
	outFile := filepath.Join(outDir, "blabla.jpg")
	defer os.RemoveAll(outDir)

	r := NewResizer(100, 100, []string{})
	err = r.resize(context.TODO(), "../../testdata/picture.jpg", outFile)
	assert.Nil(t, err)
	_, err = os.Stat(outFile)
	assert.Nil(t, err)
}

func TestResizer_NonExistingSource(t *testing.T) {
	outDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	assert.Nil(t, err)
	t.Logf("temp folder: %s", outDir)
	outFile := filepath.Join(outDir, "blabla.jpg")
	defer os.RemoveAll(outDir)

	r := NewResizer(100, 100, []string{})
	assert.NotNil(t, r.resize(context.TODO(), "../testdata/nonExisting.jpg", outFile))
}

func TestResizer_FailOnResizing(t *testing.T) {
	outDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	assert.Nil(t, err)
	t.Logf("temp folder: %s", outDir)
	defer os.RemoveAll(outDir)

	r := NewResizer(100, 100, []string{})
	err = r.resize(context.TODO(), "../../testdata/picture.jpg", "/blabliblu/aaa.jpg")
	t.Logf("error: %v", err)
	assert.NotNil(t, err)
}

func TestResizerCleanUp_Nominal(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "TestNopResizer_CleanUp")
	assert.Nil(t, err)
	defer os.Remove(f.Name())
	r := NewResizer(640, 480, []string{})
	assert.Nil(t, r.cleanup(context.TODO(), f.Name()))
	_, err = os.Stat("/path/to/whatever")
	assert.True(t, os.IsNotExist(err))
}

func TestResizerCleanUp_NonExisting(t *testing.T) {
	assert.NotNil(t, NewResizer(640, 480, []string{}).cleanup(context.TODO(), "nonExistingFile"))
}

func TestNewResizer_fallbackExt(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inExt  []string
		expExt []string
	}{
		{"nil", nil, []string{}},
		{"empty", []string{}, []string{}},
		{"single", []string{"vAl"}, []string{"val"}},
		{"multiple", []string{"vAl1", "VaL2"}, []string{"val1", "val2"}},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			r := NewResizer(100, 100, tc.inExt)
			assert.ElementsMatch(t, tc.expExt, r.fallbackExt)
		})
	}
}

func TestResizerHasToFallback(t *testing.T) {
	r := NewResizer(100, 100, []string{"ExT1", "eXt2"})
	assert.True(t, r.hasToFallback("/tmp/a.ext1"))
	assert.True(t, r.hasToFallback("/tmp/a.EXt1"))
	assert.True(t, r.hasToFallback("/tmp/a.EXT2"))
	assert.False(t, r.hasToFallback("/tmp/.ext1/a.txt"))
	assert.False(t, r.hasToFallback("/tmp/a.doc"))
}
