package browse

import (
	"context"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

const imageMimeTypePrefix = "image/"

type Task struct {
	Path string
	Info os.FileInfo
}

func BrowseImages(ctx context.Context, dirList []string, outFileChan chan Task) error {
	defer close(outFileChan)
	for _, curDir := range dirList {
		err := filepath.Walk(curDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if isPicture(path) {
					outFileChan <- Task{
						Path: path,
						Info: info,
					}
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error while browsing %v: %w", curDir, err)
		}
	}
	return nil
}

func isPicture(path string) bool {
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		log.Warn().Str("file", path).Msgf("Error while getting mime-type for %v: %v", path, err)
	}
	return err == nil && strings.HasPrefix(mime.String(), imageMimeTypePrefix)
}
