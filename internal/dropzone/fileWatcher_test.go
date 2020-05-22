package dropzone

import (
	"github.com/barasher/picdexer/conf"
	"github.com/stretchr/testify/assert"
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
	c := conf.DropzoneConf{Root: "../../testdata/"}
	f, err := NewFileWatcher(c)
	assert.Nil(t, err)

	items, err := f.Watch()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(items))

	items, err = f.Watch()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(items))
	checkHasItem(t, items, "../../testdata/nonPictureFile.txt")
	checkHasItem(t, items, "../../testdata/picture.jpg")
}
