package browse

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/internal/common"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

type Task struct {
	Path   string
	Info   os.FileInfo
	FileID string
}

type Browser struct{}

func (*Browser) Browse(ctx context.Context, dirList []string, outFileChan chan Task) error {
	defer close(outFileChan)
	for _, curDir := range dirList {
		err := filepath.Walk(curDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if isPic, key, err := common.CategorizePicture(path); err == nil {
					if isPic {
						outFileChan <- Task{
							Path:   path,
							Info:   info,
							FileID: key,
						}
					}
				} else {
					log.Warn().Str(common.LogFileIdentifier, path).Msgf("%v", err)
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
