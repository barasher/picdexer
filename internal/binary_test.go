package internal

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestNewBinaryManager_ErrorOnOpts(t *testing.T) {
	_, err := NewBinaryManager(func(*BinaryManager) error {
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
		expSuccess  bool
		expValue    int
	}{
		{"-1", -1, false, 0},
		{"0", 0, false, 0},
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
			bm, err := NewBinaryManager(BinaryManagerDoResize(tc.inW, tc.inH))
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
	bm, err := NewBinaryManager(BinaryManagerDoPush("anUrl"))
	assert.Nil(t, err)
	pusher, ok := bm.pusher.(pusher)
	assert.True(t, ok)
	assert.Equal(t, "anUrl", pusher.url)
}

type mockSubStore struct {
	resized bool
	pushed  bool
}

func (m *mockSubStore) resize(ctx context.Context, f string, o string) (string, string, error) {
	m.resized = true
	return f, filepath.Base(f), nil
}

func (m *mockSubStore) push(bin string, key string) error {
	m.pushed = true
	return nil
}

func TestStore(t *testing.T) {
	mock := &mockSubStore{}
	bm, err := NewBinaryManager()
	assert.Nil(t, err)
	bm.resizer = mock
	bm.pusher = mock

	f := "../testdata/picture.jpg"
	fInfo, err := os.Stat(f)
	assert.Nil(t, err)
	in := make(chan Task, 1)
	in <- Task{Path: f, Info: fInfo}
	close(in)

	assert.Nil(t, bm.Store(context.TODO(), in, ""))
	assert.True(t, mock.resized)
	assert.True(t, mock.pushed)
}
