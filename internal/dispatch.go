package internal

import "context"

func DispatchTasks(ctx context.Context, inFileChan chan Task, outIdxChan chan Task, outBinChan chan Task) {
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
			outIdxChan<-t
			outBinChan<-t
		}
	}
}
