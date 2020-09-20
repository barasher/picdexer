package cmd

import (
	"context"
	"fmt"
	exif "github.com/barasher/go-exiftool"
	"github.com/barasher/picdexer/conf"
	"github.com/barasher/picdexer/internal/common"
	"github.com/barasher/picdexer/internal/metadata"
	"github.com/spf13/cobra"
)

var (
	metaCmd = &cobra.Command{
		Use:   "metadata",
		Short: "Picdexer : metadata utilities",
	}

	metaSimuCmd = &cobra.Command{
		Use:   "simulate",
		Short: "Simulate metadata extraction",
		RunE:  simulateMeta,
	}

	metaDisplayCmd = &cobra.Command{
		Use:   "display",
		Short: "Display file metadata",
		RunE:  displayMeta,
	}

	metaIndexCmd = &cobra.Command{
		Use:   "index",
		Short: "Index file/directory",
		RunE:  indexMeta,
	}
)

func init() {
	// simulate
	metaSimuCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	metaSimuCmd.Flags().StringArrayVarP(&input, "dir", "d", []string{}, "Directory/File containing pictures")
	metaSimuCmd.Flags().StringVarP(&importID, "impId", "i", "", "Import identifier")
	metaSimuCmd.MarkFlagRequired("dir")
	metaCmd.AddCommand(metaSimuCmd)

	// display
	metaDisplayCmd.Flags().StringArrayVarP(&input, "dir", "d", []string{}, "File to extract")
	metaDisplayCmd.MarkFlagRequired("file")
	metaCmd.AddCommand(metaDisplayCmd)

	// index
	metaIndexCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	metaIndexCmd.Flags().StringArrayVarP(&input, "dir", "d", []string{}, "Directory/File containing pictures")
	metaIndexCmd.Flags().StringVarP(&importID, "impId", "i", "", "Import identifier")
	metaIndexCmd.MarkFlagRequired("dir")
	metaIndexCmd.MarkFlagRequired("conf")
	metaCmd.AddCommand(metaIndexCmd)

	rootCmd.AddCommand(metaCmd)
}

func extractConfigured(ctx context.Context, conf conf.Conf, inputs []string, push bool) error {
	if err := setLoggingLevel(conf.LogLevel) ; err != nil {
		return fmt.Errorf("error while configuring logging level: %w", err)
	}

	idxer, err := metadata.NewIndexer(conf.Elasticsearch)
	if err != nil {
		return fmt.Errorf("error while initializing metadata: %w", err)
	}
	defer idxer.Close()

	for _, curInput := range inputs {
		if push {
			if err := idxer.ExtractAndPushFolder(ctx, curInput); err != nil {
				return fmt.Errorf("error while extracting metadata: %w", err)
			}
		} else {
			if err := idxer.ExtractFolder(ctx, curInput); err != nil {
				return fmt.Errorf("error while extracting metadata: %w", err)
			}
		}
	}
	return nil
}

func extract(ctx context.Context, push bool) error {
	if confFile == "" {
		return fmt.Errorf("no configuration file provided")
	}
	c, err := conf.LoadConf(confFile)
	if err != nil {
		return fmt.Errorf("error while loading configuration")
	}

	return extractConfigured(ctx, c, input, push)
}

func simulateMeta(cmd *cobra.Command, args []string) error {
	ctx := common.NewContext(importID)
	return extract(ctx, false)
}

func indexMeta(cmd *cobra.Command, args []string) error {
	ctx := common.NewContext(importID)
	return extract(ctx, true)
}

func displayMeta(cmd *cobra.Command, args []string) error {
	et, err := exif.NewExiftool()
	if err != nil {
		return fmt.Errorf("error while initializing metadata extractor: %v", err)
	}
	defer et.Close()

	metas := et.ExtractMetadata(input...)
	if len(metas) != 1 {
		return fmt.Errorf("wrong metadatas count (%v)", len(metas))
	}

	if metas[0].Err != nil {
		return fmt.Errorf("Error while extracting metadatas: %v", metas[0].Err)
	}

	for k, v := range metas[0].Fields {
		fmt.Printf("%v: %v\n", k, v)
	}

	return nil
}
