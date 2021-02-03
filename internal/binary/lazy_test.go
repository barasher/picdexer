package binary

import (
	"context"
	"github.com/barasher/picdexer/internal/browse"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLazyStore(t *testing.T) {
	f := "../../testdata/picture.jpg"
	fInfo, err := os.Stat(f)
	assert.Nil(t, err)
	in := make(chan browse.Task, 2)
	in <- browse.Task{Path: f, Info: fInfo}
	in <- browse.Task{Path: f, Info: fInfo}
	close(in)

	l := LazyBinaryManager{}
	assert.Nil(t, l.Store(context.TODO(), in, ""))
}
