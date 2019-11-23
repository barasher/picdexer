package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/barasher/picdexer/internal/indexer"
	"github.com/sirupsen/logrus"
)

var indexCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump file/directory",
	Run:   Dump,
}

func init() {
	indexCmd.Flags().StringVarP(&input, "dir", "d", "", "Directory/File to index")
	indexCmd.MarkFlagRequired("dir")
	rootCmd.AddCommand(indexCmd)
}

func Dump(cmd *cobra.Command, args []string) {
	ret, err := doDump()
	if err != nil {
		logrus.Errorf("%v", err)
	}
	os.Exit(ret)
}

func doDump() (int, error) {
	opts := []func(*indexer.Indexer) error{}
	opts = append(opts, indexer.Input(input))

	idxer, err := indexer.NewIndexer(opts...)
	if err != nil {
		return retExecFailure, fmt.Errorf("error while initializing indexer: %v", err)
	}
	defer idxer.Close()

	if err := idxer.Dump(); err != nil {
		return retExecFailure, fmt.Errorf("error while dumping: %v", err)
	}

	return retOk, nil
}
