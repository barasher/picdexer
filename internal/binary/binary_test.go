package binary

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewStorer(t *testing.T) {
	var tcs = []struct {
		tcID       string
		inConf     conf.BinaryConf
		inPush     bool
		expSuccess bool
		expResizer string
		expPusher  string
	}{
		{"Nominal_1", conf.BinaryConf{}, false, true, "binary.nopResizer", "binary.nopPusher"},
		{"Nominal_2", conf.BinaryConf{Height: 1, Width: 1}, false, true, "binary.resizer", "binary.nopPusher"},
		{"Nominal_3", conf.BinaryConf{}, true, true, "binary.nopResizer", "binary.pusher"},
		{"Nominal_4", conf.BinaryConf{Height: 1, Width: 1}, true, true, "binary.resizer", "binary.pusher"},
		{"MissingDimension", conf.BinaryConf{Width: 1}, true, false, "", ""},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			s, err := NewStorer(tc.inConf, tc.inPush)
			assert.Equal(t, tc.expSuccess, err == nil)
			if tc.expSuccess {
				assert.Equal(t, tc.expResizer, reflect.TypeOf(s.resizer).String())
				assert.Equal(t, tc.expPusher, reflect.TypeOf(s.pusher).String())
			}
		})
	}
}

type resizerMock struct {
	f   string
	k   string
	err error
}

func (r resizerMock) resize(ctx context.Context, f string, o string) (string, string, error) {
	return r.f, r.k, r.err
}

type pusherMock struct {
	err error
}

func (p pusherMock) push(bin string, key string) error {
	return p.err
}

func TestStoreFolder(t *testing.T) {
	var tcs = []struct {
		tcID       string
		inResizerI resizerInterface
		inPushI    pusherInterface
	}{
		{
			"Nominal",
			resizerMock{f: "../../testdata/picture.jpg", k: "myKey"},
			pusherMock{},
		}, {
			"FailResize",
			resizerMock{err: fmt.Errorf("error")},
			pusherMock{},
		}, {
			"FailPush",
			resizerMock{f: "../../testdata/picture.jpg", k: "myKey"},
			pusherMock{err: fmt.Errorf("error")},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			s := Storer{
				resizer: tc.inResizerI,
				pusher:  tc.inPushI,
			}
			err := s.StoreFolder(context.TODO(), "../../testdata/picture.jpg", "")
			assert.Nil(t, err)
		})
	}

}
