package binary

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNopResizer_Nominal(t *testing.T) {
	r := NewNopResizer()
	outBin, outKey, err := r.resize(context.TODO(), "../../testdata/picture.jpg", "/tmp")
	assert.Nil(t, err)
	assert.Equal(t, "../../testdata/picture.jpg", outBin)
	assert.Equal(t, "ec3d25618be7af41c6824855f0f42c73_picture.jpg", outKey)
}

func TestNopResizer_NonExisting(t *testing.T) {
	r := NewNopResizer()
	_, _, err := r.resize(context.TODO(), "../testdata/nonExisting.jpg", "/tmp")
	assert.NotNil(t, err)
}

func TestNopResizerCleanUp(t *testing.T) {
	assert.Nil(t, NewNopResizer().cleanup(context.TODO(), "blabla"))
}

func TestGetOutputFilename_Nominal(t *testing.T) {
	out, err := getOutputFilename("../../testdata/picture.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "ec3d25618be7af41c6824855f0f42c73_picture.jpg", out)
}

func TestGetOutputFilename_NonExisting(t *testing.T) {
	_, err := getOutputFilename("../testdata/nonExisting.jpg")
	assert.NotNil(t, err)
}

func TestResizer_Nominal(t *testing.T) {
	outDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	assert.Nil(t, err)
	t.Logf("temp folder: %s", outDir)
	defer os.RemoveAll(outDir)

	r := NewResizer(100, 100, []string{})
	bin, key, err := r.resize(context.TODO(), "../../testdata/picture.jpg", outDir)
	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(outDir, "ec3d25618be7af41c6824855f0f42c73_picture.jpg"), bin)
	assert.Equal(t, key, "ec3d25618be7af41c6824855f0f42c73_picture.jpg")
}

func TestResizer_NonExistingSource(t *testing.T) {
	outDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	assert.Nil(t, err)
	t.Logf("temp folder: %s", outDir)
	defer os.RemoveAll(outDir)

	r := NewResizer(100, 100, []string{})
	_, _, err = r.resize(context.TODO(), "../testdata/nonExisting.jpg", outDir)
	assert.NotNil(t, err)
}

func TestResizer_FailOnResizing(t *testing.T) {
	outDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	assert.Nil(t, err)
	t.Logf("temp folder: %s", outDir)
	defer os.RemoveAll(outDir)

	r := NewResizer(100, 100, []string{})
	_, _, err = r.resize(context.TODO(), "../../testdata/picture.jpg", "/blabliblu/")
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
		tcID       string
		inExt   []string
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
