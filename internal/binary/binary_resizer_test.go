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

	r := NewResizer(100,100)
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

	r := NewResizer(100,100)
	_, _, err = r.resize(context.TODO(), "../testdata/nonExisting.jpg", outDir)
	assert.NotNil(t, err)
}

func TestResizer_FailOnResizing(t *testing.T) {
	outDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	assert.Nil(t, err)
	t.Logf("temp folder: %s", outDir)
	defer os.RemoveAll(outDir)

	r := NewResizer(100,100)
	_, _, err = r.resize(context.TODO(), "../../testdata/picture.jpg", "/blabliblu/")
	t.Logf("error: %v", err)
	assert.NotNil(t, err)
}
