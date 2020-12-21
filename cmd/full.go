package cmd

import (
	"fmt"
	"github.com/barasher/picdexer/internal/common"
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
	fullCmd.Flags().StringArrayVarP(&input, "dir", "d", []string{}, "Directory/File containing pictures")
	fullCmd.Flags().StringVarP(&importID, "impId", "i", "", "Import identifier")

	/*fullCmd.Flags().BoolVarP(&doNotExtractMetadata, "doNotExtractMetadata", "", false, "Does not extract metadata")
	fullCmd.Flags().BoolVarP(&doNotIndex, "doNotIndex", "", false, "Does not index metadata")
	fullCmd.Flags().BoolVarP(&doNotUpload, "doNotUpload", "", false, "Does not upload picture")
	fullCmd.Flags().BoolVarP(&doNotResize, "doNotResize", "", false, "Does not resize")*/

	fullCmd.MarkFlagRequired("conf")
	fullCmd.MarkFlagRequired("dir")
	rootCmd.AddCommand(fullCmd)
}

func full(cmd *cobra.Command, args []string) error {
	ctx := common.NewContext(importID)
	var c Config
	var err error
	if confFile != "" {
		if c, err = LoadConf(confFile); err != nil {
			return fmt.Errorf("error while loading configuration (%v): %w", confFile, err)
		}
	}
	return Run(ctx, c, input)
}
