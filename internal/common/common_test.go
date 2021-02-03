package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsPicture(t *testing.T) {
	assert.False(t, IsPicture("../../testdata/nonPictureFile.txt"))
	assert.True(t, IsPicture("../../testdata/picture.jpg"))
	assert.False(t, IsPicture("../../testdata/blabla"))
}

