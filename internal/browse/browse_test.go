package browse

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBrowse(t *testing.T) {
	taskChan := make(chan Task, 10)
	files := []string{}
	go func() {
		b := &Browser{}
		assert.Nil(t, b.Browse(context.Background(), []string{"../../testdata"}, taskChan))
	}()
	for cur := range taskChan {
		files = append(files, cur.Path)
	}
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "../../testdata/picture.jpg", files[0])
}
