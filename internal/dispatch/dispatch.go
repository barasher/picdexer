package dispatch

import (
	"context"
	"github.com/barasher/picdexer/internal/browse"
)

func DispatchTasks(ctx context.Context, inFileChan chan browse.Task, outIdxChan chan browse.Task, outBinChan chan browse.Task) {
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-inFileChan:
			if !ok {
				close(outIdxChan)
				close(outBinChan)
				return
			}
			outIdxChan <- t
			outBinChan <- t
		}
	}
}
