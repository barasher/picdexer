package binary

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/internal/browse"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestNewBinaryManager_ErrorOnOpts(t *testing.T) {
	_, err := NewBinaryManager(2, func(*BinaryManager) error {
		return fmt.Errorf("anError")
	})
	assert.NotNil(t, err)
}

func TestNewBinaryManager(t *testing.T) {
	var tcs = []struct {
		inTC  int
		expOk bool
	}{
		{-1, false},
		{0, false},
		{2, true},
	}
	for _, tc := range tcs {
		t.Run(strconv.Itoa(tc.inTC), func(t *testing.T) {
			bm, err := NewBinaryManager(tc.inTC)
			if tc.expOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.inTC, bm.threadCount)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestBinaryManagerDoResize(t *testing.T) {
	var tcs = []struct {
		tcID  string
		inW   int
		inH   int
		expOk bool
	}{
		{"1x2", 1, 2, true},
		{"1x0", 1, 0, false},
		{"0x1", 0, 1, false},
		{"0x0", 0, 0, false},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			bm, err := NewBinaryManager(4, BinaryManagerDoResize(tc.inW, tc.inH, []string{}))
			if tc.expOk {
				assert.Nil(t, err)
				resizer, ok := bm.resizer.(resizer)
				assert.True(t, ok)
				assert.Equal(t, "1x2", resizer.dimensions)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestBinaryManagerDoPush(t *testing.T) {
	bm, err := NewBinaryManager(4, BinaryManagerDoPush("anUrl"))
	assert.Nil(t, err)
	pusher, ok := bm.pusher.(pusher)
	assert.True(t, ok)
	assert.Equal(t, "anUrl", pusher.url)
}

type mockSubStore struct {
	resized   bool
	pushed    bool
	cleanedUp bool
}

func (m *mockSubStore) resize(ctx context.Context, from string, to string) error {
	m.resized = true
	return nil
}

func (m *mockSubStore) cleanup(ctx context.Context, f string) error {
	m.cleanedUp = true
	return nil
}

func (m *mockSubStore) push(bin string, key string) error {
	m.pushed = true
	return nil
}

func TestStore(t *testing.T) {
	mock := &mockSubStore{}
	bm, err := NewBinaryManager(4)
	assert.Nil(t, err)
	bm.resizer = mock
	bm.pusher = mock

	f := "../../testdata/picture.jpg"
	fInfo, err := os.Stat(f)
	assert.Nil(t, err)
	in := make(chan browse.Task, 1)
	in <- browse.Task{Path: f, Info: fInfo}
	close(in)

	assert.Nil(t, bm.Store(context.TODO(), in, ""))
	assert.True(t, mock.resized)
	assert.True(t, mock.pushed)
	assert.True(t, mock.cleanedUp)
}
