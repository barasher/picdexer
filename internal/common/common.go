package common

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/rs/zerolog/log"
	"strings"
)

const (
	LogFileIdentifier   = "file"
	imageMimeTypePrefix = "image/"
)

func IsPicture(path string) bool {
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		log.Warn().Str("file", path).Msgf("Error while getting mime-type for %v: %v", path, err)
	}
	return err == nil && strings.HasPrefix(mime.String(), imageMimeTypePrefix)
}
