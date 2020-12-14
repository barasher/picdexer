package internal

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBrowse(t *testing.T) {
	taskChan := make (chan Task, 10)
	files := []string{}
	go func(){
		assert.Nil(t, BrowseImages(context.Background(), "../testdata", taskChan))
	}()
	for cur := range taskChan {
		files = append(files, cur.Path)
	}
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "../testdata/picture.jpg", files[0])
}

func TestIsPicture(t *testing.T) {
	assert.False(t, isPicture("../testdata/nonPictureFile.txt"))
	assert.True(t, isPicture("../testdata/picture.jpg"))
	assert.False(t, isPicture("../testdata/blabla"))
}