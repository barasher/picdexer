package binary

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
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

func (s *Storer) StoreFolder(ctx context.Context, f string, o string) error {
	c := make(chan string, s.conf.ConversionThreads)
	go func() {
		err := filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				c <- path
			}
			return nil
		})
		close(c)
		if err != nil {
			logrus.Errorf("error while browsing folder %v: %w", f, err)
		}
	}()
	return s.StoreChannel(ctx, c, o)
}

func (s *Storer) StoreChannel(ctx context.Context, c <-chan string, o string) error {
	var dir = o
	var err error
	if o == "" {
		dir, err = ioutil.TempDir(os.TempDir(), "picdexer")
		if err != nil {
			return fmt.Errorf("error while creating temporary folder: %w", err)
		}
		defer os.RemoveAll(dir)
		log.Debug().Msgf("Resized pictures temporary folder: %v", dir)
	}

	for cur := range c {

		log.Debug().Str(fileIdentifier, cur).Msg("Resizing...")
		outBin, outKey, err := s.resizer.resize(ctx, cur, dir)
		if err != nil {
			log.Error().Str(fileIdentifier, cur).Msgf("Error while resizing: %v", err)
			continue
		}

		log.Debug().Str(fileIdentifier, cur).Str(resizedFileIdentifier, outBin).Str(keyIdentifier, outKey).Msg("Pushing...")
		err = s.pusher.push(outBin, outKey)
		if err != nil {
			log.Error().Str(fileIdentifier, cur).Str(resizedFileIdentifier, outBin).Str(keyIdentifier, outKey).Msgf("Error while pushing: %v", err)
			continue
		}

	}
	return nil
}
