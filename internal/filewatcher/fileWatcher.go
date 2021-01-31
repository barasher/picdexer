package filewatcher

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

type Item struct {
	Path string
	Info os.FileInfo
}

type FileWatcher struct {
	root       string
	OldWatched map[string]Item
}

func NewFileWatcher(rootDir string) *FileWatcher {
	fw := &FileWatcher{
		root:       rootDir,
		OldWatched: make(map[string]Item),
	}
	return fw
}

func (fw *FileWatcher) scanDir() (map[string]Item, error) {
	items := make(map[string]Item)
	err := filepath.Walk(fw.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			items[path] = Item{
				Path: path,
				Info: info,
			}
		}
		return nil
	})
	return items, err
}

func (fw *FileWatcher) Watch() ([]Item, error) {
	items, err := fw.scanDir()
	if err != nil {
		return nil, fmt.Errorf("error while scanning folder: %s", err)
	}
	res := []Item{}

	if len(fw.OldWatched) == 0 { // first watch
		fw.OldWatched = items
		return res, nil
	}

	for k, v := range items {
		oldV, found := fw.OldWatched[k]
		if found {
			log.Debug().Msgf("fileWatcher: %s: %s -> %s", v.Path, oldV.Info.ModTime(), v.Info.ModTime())
			if oldV.Info.ModTime() == v.Info.ModTime() {
				log.Debug().Msgf("fileWatcher: detected file: %v", oldV.Path)
				delete(items, k)
				res = append(res, v)
			}
		}
	}
	fw.OldWatched = items

	return res, nil
}
