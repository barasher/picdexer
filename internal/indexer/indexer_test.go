package indexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDegMinSecToDecimal(t *testing.T) {
	var tcs = []struct {
		tcID   string
		inDeg  string
		inMin  string
		inSec  string
		inLet  string
		expOk  bool
		expVal float32
	}{
		{"Unparsable deg", "bla", "1.0", "1.0", "N", false, 0},
		{"Unparsable min", "1.0", "bla", "1.0", "N", false, 0},
		{"Unparsable sec", "1.0", "1.0", "bla", "N", false, 0},
		{"Unsupported letter", "1.0", "1.0", "1.0", "Q", false, 0},
		{"Nominal decimal", "1.0", "11.0", "60.0", "N", true, 1.2},
		{"Nominal integer N", "1", "11", "60", "N", true, 1.2},
		{"Nominal integer E", "1", "11", "60", "E", true, 1.2},
		{"Nominal integer S", "1", "11", "60", "S", true, -1.2},
		{"Nominal integer W", "1", "11", "60", "W", true, -1.2},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			v, err := degMinSecToDecimal(tc.inDeg, tc.inMin, tc.inSec, tc.inLet)
			if tc.expOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.expVal, v)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestConvertGPSCoordinates(t *testing.T) {
	var tcs = []struct {
		tcID      string
		inLatLong string
		expOk     bool
		expLat    float32
		expLong   float32
	}{
		{"Nominal", `1 deg 11' 60" N, 1 deg 11' 60" W`, true, 1.2, -1.2},
		{"Unparsable latitude", `b deg 11' 60" N, 1 deg 11' 60" W`, false, 0,0},
		{"Unparsable longitude", `1 deg 11' 60" N, b deg 11' 60" W`, false, 0,0},
		{"Wrong size", `a b`, false, 0,0},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			lat, long, err := convertGPSCoordinates(tc.inLatLong)
			if tc.expOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.expLat, lat)
				assert.Equal(t, tc.expLong, long)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
