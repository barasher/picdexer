package cmd

import (
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/barasher/picdexer/internal/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	fullCmd = &cobra.Command{
		Use:   "full",
		Short: "Picdexer : indexing & storing",
		RunE:  full,
	}
)

func init() {
	// full
	fullCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	fullCmd.Flags().StringVarP(&input, "dir", "d", "", "Directory/File containing pictures")
	fullCmd.Flags().StringVarP(&importID, "impId", "i", "", "Import identifier")
	fullCmd.MarkFlagRequired("conf")
	fullCmd.MarkFlagRequired("dir")

	rootCmd.AddCommand(fullCmd)
}

func full(cmd *cobra.Command, args []string) error {
	ctx := common.NewContext(importID)

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

	log.Info().Msg("Indexing metadata...")
	if err :=  extractConfigured(ctx, c, true); err != nil {
		return fmt.Errorf("Error while indexing metadata: %w", err)
	}
	log.Info().Msg("Storing pictures...")
	if err :=  doBinConfigured(true, c, input, ""); err != nil {
		return fmt.Errorf("Error while storing pictures: %w", err)
	}
	return nil
}