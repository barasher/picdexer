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
