package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBulkSize(t *testing.T) {
	var tcs = []struct {
		tcID        string
		inConfValue int
		expValue    int
	}{
		{"-1", -1, defaultBulkSize},
		{"0", 0, defaultBulkSize},
		{"5", 5, 5},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			c := EsPusherConf{BulkSize: tc.inConfValue}
			p, err := NewEsPusher(c)
			assert.Nil(t, err)
			assert.Equal(t, tc.expValue, p.bulkSize())
		})
	}
}
