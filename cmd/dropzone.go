package cmd

import (
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/barasher/picdexer/internal/binary"
	"github.com/barasher/picdexer/internal/common"
	dropzone2 "github.com/barasher/picdexer/internal/dropzone"
	"github.com/barasher/picdexer/internal/metadata"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"time"
)

const (
	defaultFileChannelSize = 20
)

var (
	dropzoneCmd = &cobra.Command{
		Use:   "dropzone",
		Short: "Picdexer : launch dropzone",
		RunE:  dropzone,
	}
)

func init() {
	dropzoneCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	dropzoneCmd.MarkFlagRequired("conf")
	rootCmd.AddCommand(dropzoneCmd)
}

func fileChannelSize(c conf.DropzoneConf) int {
	n := c.FileChannelSize
	if n < 1 {
		n = defaultFileChannelSize
	}
	return n
}

func dropzone(cmd *cobra.Command, args []string) error {
	var c conf.Conf
	var err error
	if confFile != "" {
		if c, err = conf.LoadConf(confFile); err != nil {
			return fmt.Errorf("error while loading configuration (%v): %w", confFile, err)
		}
	}

	if err := setLoggingLevel(c.LogLevel) ; err != nil {
		return fmt.Errorf("error while configuring logging level: %w", err)
	}

	ctx := common.NewContext(importID)

	fw, err := dropzone2.NewFileWatcher(c.Dropzone)
	if err != nil {
		return fmt.Errorf("error while watching folder %v: %w", c.Dropzone.Root, err)
	}
	period, err := time.ParseDuration(c.Dropzone.Period)
	if err != nil {
		return fmt.Errorf("error while parsing watching duration (%s): %w", c.Dropzone.Period, err)
	}

	idxer, err := metadata.NewIndexer(c.Elasticsearch)
	if err != nil {
		return fmt.Errorf("error while initializing metadata: %w", err)
	}
	defer idxer.Close()

	s, err := binary.NewStorer(c.Binary, true)
	if err != nil {
		return fmt.Errorf("error while initializing storer: %w", err)
	}
	binTmpDir, err := ioutil.TempDir(os.TempDir(), "picdexer")
	if err != nil {
		return fmt.Errorf("error while creating temporary folder: %v", err)
	}

	for {
		log.Debug().Msgf("Watching iteration...")

		items, err := fw.Watch()
		if err != nil {
			log.Error().Msgf("Error while watching: %s", err)
		}
		if len(items) > 0 {
			metaTasks := make(chan metadata.ExtractTask, fileChannelSize(c.Dropzone))
			go func() {
				for _, cur := range items {
					if common.IsPicture(cur.Path) {
						metaTasks <- metadata.ExtractTask{Path: cur.Path, Info: cur.Info}
					}
				}
				close(metaTasks)
			}()
			err = idxer.ExtractAndPushTasks(ctx, metaTasks)
			if err != nil {
				log.Error().Msgf("Error while extracting tasks: %v", err)
			}

			binTasks := make(chan string, fileChannelSize(c.Dropzone))
			go func() {
				for _, cur := range items {
					if common.IsPicture(cur.Path) {
						binTasks <- cur.Path
					}
				}
				close(binTasks)
			}()
			s.StoreChannel(ctx, binTasks, binTmpDir)

			for _, cur := range items {
				if err := os.Remove(cur.Path); err != nil {
					log.Error().Msgf("Error while deleting %v: %v", cur.Path, err)
				}
			}
		}

		time.Sleep(period)

		// TODO filter file type
	}

	return nil
}
