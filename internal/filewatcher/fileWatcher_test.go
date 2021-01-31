package filewatcher

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func checkHasItem(t *testing.T, items []Item, path string) {
	for _, cur := range items {
		if cur.Path == path {
			return
		}
	}
	assert.Fail(t, "%v not found", path)
}

func TestWatch(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "TestWatch")
	assert.Nil(t, err)
	t.Logf("tmpDir: %v", tmpDir)
	defer os.RemoveAll(tmpDir)

	fw := NewFileWatcher(tmpDir)

	items, err := fw.Watch()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(items))

	dest := tmpDir + "/picture.jpg"
	assert.Nil(t, copy("../../testdata/picture.jpg", dest))

	items, err = fw.Watch()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(items))

	items, err = fw.Watch()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(items))
	checkHasItem(t, items, dest)
}

func copy(src, dest string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, input, 0644)
	if err != nil {
		return err
	}

	return nil
}
