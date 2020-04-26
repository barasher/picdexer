package cmd

import (
	"fmt"
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
	fullCmd.MarkFlagRequired("conf")
	fullCmd.MarkFlagRequired("dir")

	rootCmd.AddCommand(fullCmd)
}

func full(cmd *cobra.Command, args []string) error {
	log.Info().Msg("Indexing metadata...")
	if err :=  extract(true); err != nil {
		return fmt.Errorf("Error while indexing metadata: %w", err)
	}
	log.Info().Msg("Storing pictures...")
	if err :=  doBin(true); err != nil {
		return fmt.Errorf("Error while storing pictures: %w", err)
	}
	return nil
}