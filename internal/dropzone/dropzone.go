package dropzone

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/rs/zerolog/log"
	fsevents "github.com/tywkeene/go-fsevents"
)

var notifMask uint32 = fsevents.CloseWrite | fsevents.DirCreatedEvent | fsevents.DirRemovedEvent

/*type createdFileHandler struct {
	fileChan chan string
}

func newCreatedFileHandler(c chan string) *createdFileHandler {
	return &createdFileHandler{fileChan: c}
}

func (c *createdFileHandler) Handle(w *fsevents.Watcher, event *fsevents.FsEvent) error {
	s, _ := os.Stat(event.Path) // FIXME
	log.Info().Msgf("file %v ? %v", event.Path, !s.IsDir())
	if !  s.IsDir() {
		c.fileChan <- event.Path
	}
	return nil
}

func (*createdFileHandler) Check(e *fsevents.FsEvent) bool {
	return e.RawEvent.Mask&fsevents.CloseWrite > 0
}

func (*createdFileHandler) GetMask() uint32 {
	return fsevents.CloseWrite
}

type createdFolder struct{}

func (c *createdFolder) Handle(w *fsevents.Watcher, event *fsevents.FsEvent) error {
	s, _ := os.Stat(event.Path) // FIXME
	if s.IsDir() {
		return nil
		log.Debug().Msgf("Directory created: %s", event.Path)
		d, err := w.AddDescriptor(event.Path, notifMask)
		if err != nil {
			return fmt.Errorf("error while specifying notification mask on folder %v: %w", event.Path, err)
		}
		if err := d.Start(); err != nil {
			return fmt.Errorf("error while starting notifications on folder %v: %w", event.Path, err)
		}
	}
	return nil
}

func (*createdFolder) Check(event *fsevents.FsEvent) bool {
	return event.RawEvent.Mask&fsevents.Create > 0
}

func (*createdFolder) GetMask() uint32 {
	return fsevents.DirCreatedEvent
}

type deletedFolder struct{}

func (c *deletedFolder) Handle(w *fsevents.Watcher, event *fsevents.FsEvent) error {
	log.Debug().Msgf("Directory removed: %s", event.Path)
	if err := w.RemoveDescriptor(event.Path); err != nil {
		return fmt.Errorf("error while deactivating notifications on folder %v: %w", event.Path, err)
	}
	return nil
}

func (*deletedFolder) Check(event *fsevents.FsEvent) bool {
	return event.IsDirRemoved()
}

func (*deletedFolder) GetMask() uint32 {
	return fsevents.DirRemovedEvent
}*/

type Watcher struct {
	FileChan chan string
	ErrChan  chan error
	cancel   context.CancelFunc
}

func NewWatcher(ctx context.Context, c conf.DropzoneConf) (*Watcher, error) {

	w, err := fsevents.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error while creating watcher: %w", err)
	}
	d, err := w.AddDescriptor(c.Root, notifMask)
	if err != nil {
		return nil, fmt.Errorf("error while adding descriptor on the root folder: %w", err)
	}
	if err := d.Start(); err != nil {
		return nil, fmt.Errorf("error while starting descriptor on the root folder: %w", err)
	}

	go w.Watch()
	wa := &Watcher{}
	wa.ErrChan = w.Errors
	wa.FileChan = make(chan string, 50)
	cctx, cancel := context.WithCancel(ctx)
	wa.cancel = cancel

	go func() {
		for {
			select {
			case <-cctx.Done():
				log.Info().Msgf("Stopping loop")
				close(wa.FileChan)
				return
			case event := <-w.Events:
				switch {
				case event.IsDirCreated():
					log.Info().Msgf("Folder created: %s", event.Path)
					d, err := w.AddDescriptor(event.Path, notifMask)
					if err != nil {
						log.Error().Msgf("Error while adding descriptor for path %s: %s\n", event.Path, err)
						break
					}
					if err := d.Start(); err != nil {
						log.Error().Msgf("Error while starting descriptor for path %s: %s\n", event.Path, err)
						break
					}
				case event.IsDirRemoved():
					log.Info().Msgf("Folder removed: %s", event.Path)
					if err := w.RemoveDescriptor(event.Path); err != nil {
						log.Error().Msgf("Error while removing descriptor for path %s: %s\n", event.Path, err)
						break
					}
				case event.IsFileCreated():
					log.Info().Msgf("File created: %s", event.Path)
					wa.FileChan <- event.Path
				}
			}
		}
	}()

	return wa, nil

}

func (w *Watcher) Stop() {
	w.cancel()
}
