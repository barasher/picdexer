package binary

import (
	"context"
	"github.com/barasher/picdexer/internal/browse"
)

type LazyBinaryManager struct {
}

func (LazyBinaryManager) Store(ctx context.Context, inTaskChan chan browse.Task, outDir string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-inTaskChan:
			if !ok {
				return nil
			}
		}
	}
	return nil
}
