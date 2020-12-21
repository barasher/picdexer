package cmd

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/internal/binary"
	"github.com/barasher/picdexer/internal/browse"
	"github.com/barasher/picdexer/internal/common"
	"github.com/barasher/picdexer/internal/dispatch"
	"github.com/barasher/picdexer/internal/elasticsearch"
	"github.com/barasher/picdexer/internal/metadata"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"sync"
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

const (
	defaultMetadataThreadCount = 4
	defaultEsBulkSize          = 30
	defaultBinaryThreadCount   = 4
)

func max(v1, v2 int) int {
	if v1 < v2 {
		return v2
	}
	return v1
}

type MetadataExtractorInterface interface {
	Close() error
	ExtractMetadata(ctx context.Context, inTaskChan chan browse.Task, outPicMetaChan chan metadata.PictureMetadata) error
}

type BinaryManagerInterface interface {
	Store(ctx context.Context, inTaskChan chan browse.Task, outDir string) error
}

type EsPusherInterface interface {
	Push(ctx context.Context, inEsDocChan chan elasticsearch.EsDoc) error
}

func buildMetadataExtractor(c Config) (MetadataExtractorInterface, int, error) {
	tc := c.Elasticsearch.ThreadCount
	if tc == 0 {
		tc = defaultMetadataThreadCount
	}
	me, err := metadata.NewMetadataExtractor(tc)
	return me, tc, err
}

func  buildEsPusher(c Config) (EsPusherInterface, error) {
	bs := c.Elasticsearch.BulkSize
	if bs == 0 {
		bs = defaultEsBulkSize
	}
	return elasticsearch.NewEsPusher(bs, elasticsearch.EsUrl(c.Elasticsearch.Url))
}

func buildBinaryManager(c Config) (BinaryManagerInterface, int, error) {
	opts := []func(manager *binary.BinaryManager) error{}
	if c.Binary.Url != "" {
		opts = append(opts, binary.BinaryManagerDoPush(c.Binary.Url))
	}
	if c.Binary.Width != 0 && c.Binary.Height != 0 {
		opts = append(opts, binary.BinaryManagerDoResize(c.Binary.Width, c.Binary.Height))
	}
	tc := c.Binary.ThreadCount
	if tc == 0 {
		tc = defaultBinaryThreadCount
	}
	bm, err := binary.NewBinaryManager(tc, opts...)
	return bm, tc, err
}

func Run(ctx context.Context, c Config, input []string) error {
	metadataExtractor, metc, err := buildMetadataExtractor(c)
	if err != nil {
		return fmt.Errorf("error while building MetadataExtractor: %w", err)
	}
	defer metadataExtractor.Close()
	binaryManager, bmtc, err := buildBinaryManager(c)
	if err != nil {
		return fmt.Errorf("error while building BinaryManager: %w", err)
	}
	esPusher, err := buildEsPusher(c)
	if err != nil {
		return fmt.Errorf("error while building EsPusher: %w", err)
	}
	browseChan := make(chan browse.Task, max(metc, bmtc))
	binToPushChan := make(chan browse.Task, bmtc)
	metaToExtractChan := make(chan browse.Task, metc)
	metaToConvertChan := make(chan metadata.PictureMetadata, metc)
	docToPushChan := make(chan elasticsearch.EsDoc, metc)

	wg := sync.WaitGroup{}
	wg.Add(5)

	go func() { // push to es
		if err := esPusher.Push(ctx, docToPushChan); err != nil {
			log.Error().Msgf("Error while pushing to Elasticsearch: %v", err)
		}
		wg.Done()
	}()

	go func() { // metadata to doc
		if err := elasticsearch.ConvertMetadataToEsDoc(ctx, metaToConvertChan, docToPushChan); err != nil {
			log.Error().Msgf("Error while converting metadata to Elasticsearch documents: %v", err)
		}
		wg.Done()
	}()

	go func() { // task to metadata
		if err := metadataExtractor.ExtractMetadata(ctx, metaToExtractChan, metaToConvertChan); err != nil {
			log.Error().Msgf("Error while extracting metadata: %v", err)
		}
		wg.Done()
	}()

	go func() { // task to binary upload
		if err := binaryManager.Store(ctx, binToPushChan, c.Binary.WorkingDir); err != nil {
			log.Error().Msgf("Error while pushing to FileServer: %v", err)
		}
		wg.Done()
	}()

	go func() { // dispatch
		dispatch.DispatchTasks(ctx, browseChan, binToPushChan, metaToExtractChan)
		wg.Done()
	}()

	// browse
	if err := browse.BrowseImages(ctx, input, browseChan); err != nil {
		log.Error().Msgf("Error while browsing input folder: %v", err)
	}

	wg.Wait()

	return nil
}