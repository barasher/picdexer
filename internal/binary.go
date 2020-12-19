package internal

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"sync"
)

const defaultBinaryManagerThreadCount = 4

type BinaryManager struct {
	threadCount int
	resizer resizerInterface
	pusher pusherInterface
}

func NewBinaryManager(opts ...func(*BinaryManager)error) (*BinaryManager, error){
	bm := &BinaryManager{
		threadCount:defaultBinaryManagerThreadCount,
		resizer: NewNopResizer(),
		pusher: NewNopPusher(),
	}
	for _, cur := range opts {
		if err := cur(bm) ; err != nil {
			return nil, fmt.Errorf("error while creating EsPusher: %w", err)
		}
	}
	return bm, nil
}

func BinaryManagerThreadCount(c int) func(*BinaryManager) error {
	return func(bm *BinaryManager) error {
		if c <= 0 {
			return fmt.Errorf("wrong thread count value (%v), must be > 0", c)
		}
		bm.threadCount = c
		return nil
	}
}

func BinaryManagerDoResize(w, h int) func(*BinaryManager) error {
	return func(bm *BinaryManager) error {
		if w==0 || h == 0 {
			return fmt.Errorf("neither width (%v) nor height (%v) can equals 0", w, h)
		}
		bm.resizer=NewResizer(w, h)
		return nil
	}
}

func BinaryManagerDoPush(url string) func(*BinaryManager) error {
	return func(bm *BinaryManager) error {
		bm.pusher=NewPusher(url)
		return nil
	}
}

func (bm *BinaryManager) Store(ctx context.Context, inTaskChan chan Task, outDir string) error {
	var dir = outDir
	var err error
	if dir == "" {
		dir, err = ioutil.TempDir(os.TempDir(), "picdexer")
		if err != nil {
			return fmt.Errorf("error while creating temporary folder: %w", err)
		}
		defer os.RemoveAll(dir)
		log.Debug().Msgf("Resized pictures temporary folder: %v", dir)
	}

	wg := sync.WaitGroup{}
	wg.Add(bm.threadCount)
	for i := 0; i < bm.threadCount; i++ {
		go func(goRoutineId int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case cur, ok := <-inTaskChan:
					if !ok {
						return
					}

					log.Info().Str(logFileIdentifier, cur.Path).Msg("Resizing...")
					outBin, outKey, err := bm.resizer.resize(ctx, cur.Path, dir)
					if err != nil {
						log.Error().Str(logFileIdentifier, cur.Path).Msgf("Error while resizing: %v", err)
						continue
					}

					log.Info().Str(logFileIdentifier, cur.Path).Str(resizedFileIdentifier, outBin).Str(logFileIdentifier, outKey).Msg("Pushing...")
					err = bm.pusher.push(outBin, outKey)
					if err != nil {
						log.Error().Str(logFileIdentifier, cur.Path).Str(resizedFileIdentifier, outBin).Str(logFileIdentifier, outKey).Msgf("Error while pushing: %v", err)
						continue
					}

				}
			}
		}(i)
	}
	wg.Wait()
	return nil
}