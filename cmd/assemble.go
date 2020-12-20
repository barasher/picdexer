package cmd

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/internal"
	"github.com/rs/zerolog/log"
	"sync"
)

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

func buildMetadataExtractor(c Config) (*internal.MetadataExtractor, int, error) {
	tc := c.Elasticsearch.ThreadCount
	if tc == 0 {
		tc = defaultMetadataThreadCount
	}
	me, err := internal.NewMetadataExtractor(tc)
	return me, tc, err
}

func  buildEsPusher(c Config) (*internal.EsPusher, error) {
	bs := c.Elasticsearch.BulkSize
	if bs == 0 {
		bs = defaultEsBulkSize
	}
	return internal.NewEsPusher(bs, internal.EsUrl(c.Elasticsearch.Url))
}

func buildBinaryManager(c Config) (*internal.BinaryManager, int, error) {
	opts := []func(manager *internal.BinaryManager) error{}
	if c.Binary.Url != "" {
		opts = append(opts, internal.BinaryManagerDoPush(c.Binary.Url))
	}
	if c.Binary.Width != 0 && c.Binary.Height != 0 {
		opts = append(opts, internal.BinaryManagerDoResize(c.Binary.Width, c.Binary.Height))
	}
	tc := c.Binary.ThreadCount
	if tc == 0 {
		tc = defaultBinaryThreadCount
	}
	bm, err := internal.NewBinaryManager(tc, opts...)
	return bm, tc, err
}

func Run(c Config, input string) error {
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
	browseChan := make(chan internal.Task, max(metc, bmtc))
	binToPushChan := make(chan internal.Task, bmtc)
	metaToExtractChan := make(chan internal.Task, metc)
	metaToConvertChan := make(chan internal.PictureMetadata, metc)
	docToPushChan := make(chan internal.EsDoc, metc)
	ctx := context.Background()

	wg := sync.WaitGroup{}
	wg.Add(5)

	go func() { // push to es
		if err := esPusher.Push(ctx, docToPushChan); err != nil {
			log.Error().Msgf("Error while pushing to Elasticsearch: %v", err)
		}
		wg.Done()
	}()

	go func() { // metadata to doc
		if err := internal.ConvertMetadataToEsDoc(ctx, metaToConvertChan, docToPushChan); err != nil {
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
		internal.DispatchTasks(ctx, browseChan, binToPushChan, metaToExtractChan)
		wg.Done()
	}()

	// browse
	if err := internal.BrowseImages(ctx, input, browseChan); err != nil {
		log.Error().Msgf("Error while browsing input folder: %v", err)
	}

	wg.Wait()

	return nil
}
