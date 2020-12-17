package internal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBinaryManager_ErrorOnOpts(t *testing.T) {
	_, err := NewBinaryManager(func(*BinaryManager)error {
		return fmt.Errorf("anError")
	})
	assert.NotNil(t, err)
}


func TestNewBinaryManager_Defaults(t *testing.T) {
	bm, err := NewBinaryManager()
	assert.Nil(t, err)
	assert.Equal(t, defaultBinaryManagerThreadCount, bm.threadCount)
}

func TestBinaryManagerThreadCount(t *testing.T) {
	var tcs = []struct {
		tcID        string
		inConfValue int
		expSuccess bool
		expValue    int
	}{
		{"-1", -1, false, 0},
		{"0", 0,  false, 0},
		{"5", 5, true, 5},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			bm, err := NewBinaryManager(BinaryManagerThreadCount(tc.inConfValue))
			if tc.expSuccess {
				assert.Nil(t, err)
				assert.Equal(t, tc.expValue, bm.threadCount)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}