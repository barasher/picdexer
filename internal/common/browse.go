package common

import (
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"github.com/gabriel-vasile/mimetype"
	"strings"
)

const imageMimeTypePrefix = "image/"

func BrowseImages(d string, f func (string, os.FileInfo)) error {
	return filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			mime, err := mimetype.DetectFile(path)
			if err != nil {
				log.Warn().Str("file", path).Msgf("error while getting mime-type: %v", err)
				return nil
			}
			if strings.HasPrefix(mime.String(), imageMimeTypePrefix) {
				f(path, info)
			}
		}
		return nil
	})
}
