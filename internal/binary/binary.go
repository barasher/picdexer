package binary

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/barasher/picdexer/internal/common"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"sync"
)

const fileIdentifier = "file"
const resizedFileIdentifier = "resizedFile"
const keyIdentifier = "key"

type Storer struct {
	conf    conf.BinaryConf
	resizer resizerInterface
	pusher  pusherInterface
}

func NewStorer(c conf.BinaryConf, push bool) (*Storer, error) {
	s := &Storer{conf: c}

	switch {
	case c.Width > 0 && c.Height > 0:
		s.resizer = NewResizer(c)
	case c.Width == 0 && c.Height == 0:
		s.resizer = NewNopResizer()
	default:
		return s, fmt.Errorf("wrong width (%w) & height (%v) couple", c.Width, c.Height)
	}

	s.pusher = NewNopPusher()
	if push {
		s.pusher = NewPusher(c)
	}

	return s, nil
}

func (s *Storer) StoreFolder(ctx context.Context, f string, o string) {
	c := make(chan string, s.conf.Threads*2)
	go func() {
		err := common.BrowseImages(f, func(path string, info os.FileInfo) {
			c <- path
		})
		close(c)
		if err != nil {
			log.Error().Msgf("error while browsing folder %v: %v", f, err)
		}
	}()
	s.StoreChannel(ctx, c, o)
}

func (s *Storer) StoreChannel(ctx context.Context, c <-chan string, o string) {
	wg := &sync.WaitGroup{}
	wg.Add(s.conf.Threads)
	for i := 0; i < s.conf.Threads; i++ {
		cur := i
		go s.storeChannel(ctx, cur, c, o, wg)
	}
	wg.Wait()
}

func (s *Storer) storeChannel(ctx context.Context, threadId int, c <-chan string, o string, wg *sync.WaitGroup) {
	defer wg.Done()

	subLog := log.With().Int("threadId", threadId).Logger()

	var dir = o
	var err error
	if o == "" {
		dir, err = ioutil.TempDir(os.TempDir(), "picdexer")
		if err != nil {
			subLog.Error().Msgf("error while creating temporary folder: %w", err)
			return
		}
		defer os.RemoveAll(dir)
		subLog.Debug().Msgf("Resized pictures temporary folder: %v", dir)
	}

	for cur := range c {

		subLog.Debug().Str(fileIdentifier, cur).Msg("Resizing...")
		outBin, outKey, err := s.resizer.resize(ctx, cur, dir)
		if err != nil {
			subLog.Error().Str(fileIdentifier, cur).Msgf("Error while resizing: %v", err)
			continue
		}

		subLog.Debug().Str(fileIdentifier, cur).Str(resizedFileIdentifier, outBin).Str(keyIdentifier, outKey).Msg("Pushing...")
		err = s.pusher.push(outBin, outKey)
		if err != nil {
			subLog.Error().Str(fileIdentifier, cur).Str(resizedFileIdentifier, outBin).Str(keyIdentifier, outKey).Msgf("Error while pushing: %v", err)
			continue
		}

	}
}
