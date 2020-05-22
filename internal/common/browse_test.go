package common

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBrowse(t *testing.T) {
	browsed := []string{}
	BrowseImages("../../testdata", func(s string, f os.FileInfo) {
		t.Logf("browsed %v", s)
		browsed = append(browsed, s)
	})
	assert.Equal(t, 1, len(browsed))
	assert.Equal(t, "../../testdata/picture.jpg", browsed[0])
}

func TestIsPicture(t *testing.T) {
	assert.False(t, IsPicture("../../testdata/nonPictureFile.txt"))
	assert.True(t, IsPicture("../../testdata/picture.jpg"))
	assert.False(t, IsPicture("../../testdata/blabla"))
}