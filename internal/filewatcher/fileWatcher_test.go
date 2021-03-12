package filewatcher

import (
	"github.com/stretchr/testify/assert"
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
	tmpDir, err := os.MkdirTemp("/tmp", "TestWatch")
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
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dest, input, 0644)
	if err != nil {
		return err
	}

	return nil
}
