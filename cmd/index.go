package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/barasher/picdexer/internal/indexer"
	"github.com/sirupsen/logrus"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index file/directory",
	Run:   Index,
}

func init() {
	indexCmd.Flags().StringVarP(&input, "dir", "d", "", "Directory/File to index")
	indexCmd.Flags().StringVarP(&esUrl, "es", "e", "", "Elasticsearch URL (ex : http://1.2.3.4:9001/)")
	indexCmd.MarkFlagRequired("dir")
	indexCmd.MarkFlagRequired("es")
	rootCmd.AddCommand(indexCmd)
}

func Index(cmd *cobra.Command, args []string) {
	ret, err := doIndex()
	if err != nil {
		logrus.Errorf("%v", err)
	}
	os.Exit(ret)
}

func doIndex() (int, error) {
	opts := []func(*indexer.Indexer) error{}
	opts = append(opts, indexer.Input(input))

	ctx := indexer.BuildContext(importID)
	idxer, err := indexer.NewIndexer(opts...)
	if err != nil {
		return retExecFailure, fmt.Errorf("error while initializing indexer: %v", err)
	}
	defer idxer.Close()

	var buffer bytes.Buffer
	if err := idxer.Dump(ctx, &buffer); err != nil {
		return retExecFailure, fmt.Errorf("error while dumping: %v", err)
	}
	if err := idxer.Push(ctx, esUrl, &buffer); err != nil {
		return retExecFailure, fmt.Errorf("error while pushing: %v", err)
	}

	return retOk, nil
}
