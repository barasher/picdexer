package cmd

import (
	"bytes"
	"fmt"
	exif "github.com/barasher/go-exiftool"
	"github.com/barasher/picdexer/conf"
	"github.com/barasher/picdexer/internal/indexer"
	"github.com/spf13/cobra"
	"os"
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
	metaSimuCmd.Flags().StringVarP(&input, "dir", "d", "", "Directory/File to index")
	metaSimuCmd.Flags().StringVarP(&importID, "impId", "i", "", "Import identifier")
	metaSimuCmd.MarkFlagRequired("dir")
	metaCmd.AddCommand(metaSimuCmd)

	// display
	metaDisplayCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	metaDisplayCmd.Flags().StringVarP(&input, "file", "f", "", "File to extract")
	metaDisplayCmd.MarkFlagRequired("file")
	metaCmd.AddCommand(metaDisplayCmd)

	// index
	metaIndexCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	metaIndexCmd.Flags().StringVarP(&input, "dir", "d", "", "Directory/File to index")
	metaIndexCmd.MarkFlagRequired("dir")
	metaCmd.AddCommand(metaIndexCmd)

	rootCmd.AddCommand(metaCmd)
}

func simulateMeta(cmd *cobra.Command, args []string) error {
	opts := []func(*indexer.Indexer) error{}
	opts = append(opts, indexer.Input(input))
	if confFile != "" {
		conf, err := conf.LoadConf(confFile)
		if err != nil {
			return err
		}
		opts = append(opts, indexer.WithConfiguration(conf))
	}

	ctx := indexer.BuildContext(importID)
	idxer, err := indexer.NewIndexer(opts...)
	if err != nil {
		return fmt.Errorf("error while initializing indexer: %w", err)
	}
	defer idxer.Close()

	if err := idxer.Dump(ctx, os.Stdout); err != nil {
		return fmt.Errorf("error while dumping: %w", err)
	}

	return nil
}

func displayMeta(cmd *cobra.Command, args []string) error {
	et, err := exif.NewExiftool()
	if err != nil {
		return fmt.Errorf("error while initializing metadata extractor: %v", err)
	}
	defer et.Close()

	metas := et.ExtractMetadata(input)
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

func indexMeta(cmd *cobra.Command, args []string) error {
	opts := []func(*indexer.Indexer) error{}
	opts = append(opts, indexer.Input(input))
	if confFile != "" {
		conf, err := conf.LoadConf(confFile)
		if err != nil {
			return err
		}
		opts = append(opts, indexer.WithConfiguration(conf))
	}

	ctx := indexer.BuildContext(importID)
	idxer, err := indexer.NewIndexer(opts...)
	if err != nil {
		return fmt.Errorf("error while initializing indexer: %v", err)
	}
	defer idxer.Close()

	var buffer bytes.Buffer
	if err := idxer.Dump(ctx, &buffer); err != nil {
		return fmt.Errorf("error while dumping: %v", err)
	}
	if err := idxer.Push(ctx, &buffer); err != nil {
		return fmt.Errorf("error while pushing: %v", err)
	}

	return nil
}
