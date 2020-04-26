package cmd

import (
	"fmt"
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
	metaSimuCmd.Flags().StringVarP(&importID, "impId", "i", "", "Import identifier")
	fullCmd.MarkFlagRequired("conf")
	fullCmd.MarkFlagRequired("dir")

	rootCmd.AddCommand(fullCmd)
}

func full(cmd *cobra.Command, args []string) error {
	ctx := common.NewContext(importID)
	log.Info().Msg("Indexing metadata...")
	if err :=  extract(ctx, true); err != nil {
		return fmt.Errorf("Error while indexing metadata: %w", err)
	}
	log.Info().Msg("Storing pictures...")
	if err :=  doBin(true); err != nil {
		return fmt.Errorf("Error while storing pictures: %w", err)
	}
	return nil
}