package cmd

import (
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
	indexCmd.MarkFlagRequired("dir")
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

	idxer, err := indexer.NewIndexer(opts...)
	if err != nil {
		return retExecFailure, fmt.Errorf("error while initializing indexer: %v", err)
	}
	defer idxer.Close()

	if err := idxer.Index(); err != nil {
		return retExecFailure, fmt.Errorf("error while indexing: %v", err)
	}

	return retOk, nil
}
