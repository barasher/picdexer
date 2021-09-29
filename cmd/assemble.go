package cmd

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/internal/binary"
	"github.com/barasher/picdexer/internal/browse"
	"github.com/barasher/picdexer/internal/dispatch"
	"github.com/barasher/picdexer/internal/elasticsearch"
	"github.com/barasher/picdexer/internal/metadata"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

const (
	defaultMetadataThreadCount = 4
	defaultEsBulkSize          = 30
	defaultBinaryThreadCount   = 4
	dateFormat                 = "2006:01:02"
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
	ConvertMetadataToEsDoc(ctx context.Context, in chan metadata.PictureMetadata, out chan elasticsearch.EsDoc) error
}

type BrowserInterface interface {
	Browse(ctx context.Context, dirList []string, outFileChan chan browse.Task) error
}

func buildMetadataExtractor(c Config) (MetadataExtractorInterface, int, error) {
	tc := c.Elasticsearch.ThreadCount
	if tc == 0 {
		tc = defaultMetadataThreadCount
	}
	me, err := metadata.NewMetadataExtractor(tc)
	return me, tc, err
}

func buildEsPusher(c Config) (EsPusherInterface, error) {
	bs := c.Elasticsearch.BulkSize
	if bs == 0 {
		bs = defaultEsBulkSize
	}
	var opts []func(*elasticsearch.EsPusher) error
	opts = append(opts, elasticsearch.EsUrl(c.Elasticsearch.Url))
	for k, d := range c.Elasticsearch.SyncOnDate {
		parsedD, err := time.Parse(dateFormat, d)
		if err != nil {
			return nil, fmt.Errorf("syncOnDate : error while parsing date %v", d)
		}
		opts = append(opts, elasticsearch.SyncOnDate(k, parsedD))
	}
	return elasticsearch.NewEsPusher(bs, opts...)
}

func buildBinaryManager(c Config) (BinaryManagerInterface, int, error) {
	if c.Binary.Url == "" { // lazy
		return binary.LazyBinaryManager{}, 1, nil
	}

	opts := []func(manager *binary.BinaryManager) error{}
	if c.Binary.Url != "" {
		opts = append(opts, binary.BinaryManagerDoPush(c.Binary.Url))
	}
	if c.Binary.Width != 0 && c.Binary.Height != 0 {
		opts = append(opts, binary.BinaryManagerDoResize(c.Binary.Width, c.Binary.Height, c.Binary.UsePreviewForExtensions))
	}
	tc := c.Binary.ThreadCount
	if tc == 0 {
		tc = defaultBinaryThreadCount
	}
	bm, err := binary.NewBinaryManager(tc, opts...)
	return bm, tc, err
}

func buildBrowser(c Config) BrowserInterface {
	return &browse.Browser{}
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
		if err := esPusher.ConvertMetadataToEsDoc(ctx, metaToConvertChan, docToPushChan); err != nil {
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
	if err := buildBrowser(c).Browse(ctx, input, browseChan); err != nil {
		log.Error().Msgf("Error while browsing input folder: %v", err)
	}

	wg.Wait()

	return nil
}
