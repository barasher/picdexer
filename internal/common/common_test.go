package common

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)



func TestGetMimeType(t *testing.T) {
	var tcs = []struct {
		tcID       string
		inPath     string
		expSuccess bool
		expMime      string
	}{
		{"txt", "../../testdata/nonPictureFile.txt", true, "text/plain"},
		{"jpg", "../../testdata/picture.jpg", true, "image/jpeg"},
		{"nonExisting", "../../testdata/blabla", false, ""},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			mime, err := getMimeType(tc.inPath)
			if tc.expSuccess {
				assert.Nil(t, err)
				assert.Truef(t, strings.HasPrefix(mime, tc.expMime), "expected prefix: %v, got: %v", tc.expMime, mime)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestHash(t *testing.T) {
	var tcs = []struct {
		tcID       string
		inPath     string
		expSuccess bool
		expHash      string
	}{
		{"txt", "../../testdata/nonPictureFile.txt", true, "d41d8cd98f00b204e9800998ecf8427e"},
		{"jpg", "../../testdata/picture.jpg", true, "ec3d25618be7af41c6824855f0f42c73"},
		{"nonExisting", "../../testdata/blabla", false, ""},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			h, err := hash(tc.inPath)
			if tc.expSuccess {
				assert.Nil(t, err)
				assert.Equal(t, tc.expHash, h)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestCategorizePicture(t *testing.T) {
	var tcs = []struct {
		tcID       string
		inPath     string
		expSuccess bool
		expIsPicture      bool
		expKey string
	}{
		{"txt", "../../testdata/nonPictureFile.txt", true, false, ""},
		{"jpg", "../../testdata/picture.jpg", true, true, "ec3d25618be7af41c6824855f0f42c73_picture.jpg"},
		{"nonExisting", "../../testdata/blabla", false, false, ""},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			isPicture, key, err := CategorizePicture(tc.inPath)
			if tc.expSuccess {
				assert.Nil(t, err)
				assert.Equal(t, tc.expIsPicture, isPicture)
				if tc.expIsPicture {
					assert.Equal(t, tc.expKey, key)
				}
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}