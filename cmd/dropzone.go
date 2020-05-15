package cmd

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/conf"
	dropzone2 "github.com/barasher/picdexer/internal/dropzone"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	dropzoneCmd = &cobra.Command{
		Use:   "dropzone",
		Short: "Picdexer : launch dropzone",
		RunE:  dropzone,
	}
)

func init() {
	// full
	dropzoneCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	dropzoneCmd.MarkFlagRequired("conf")

	rootCmd.AddCommand(dropzoneCmd)
}

func dropzone(cmd *cobra.Command, args []string) error {
	var c conf.Conf
	var err error
	if confFile != "" {
		if c, err = conf.LoadConf(confFile); err != nil {
			return fmt.Errorf("error while loading configuration (%v): %w", confFile, err)
		}
	}

	w, err := dropzone2.NewWatcher(context.Background(), c.Dropzone)
	if err != nil {
		return fmt.Errorf("error while watching folder %v: %w", c.Dropzone.Root, err)
	}



	for {
		select {
		case event := <-w.FileChan:
			log.Info().Msgf("Consuming event %v", event)
		case err = <-w.ErrChan:
			log.Error().Msgf("Consuming error %v", err)
		}
	}

	return nil
}
