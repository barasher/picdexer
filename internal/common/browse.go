package common

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

const imageMimeTypePrefix = "image/"

func BrowseImages(d string, f func (string, os.FileInfo)) error {
	return filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if IsPicture(path) {
				f(path, info)
			}
		}
		return nil
	})
}

func IsPicture(path string) bool{
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		log.Warn().Str("file", path).Msgf("Error while getting mime-type for %v: %v", path, err)
	}
	return err == nil &&  strings.HasPrefix(mime.String(), imageMimeTypePrefix)
}