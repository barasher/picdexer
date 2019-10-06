package main

import (
	"flag"
"os"
	
	"github.com/barasher/picdexer/internal/indexer"
	"github.com/sirupsen/logrus"
)

const (
	retOk          int = 0
	retConfFailure int = 1
	retExecFailure int = 2
)

func main() {
	os.Exit(doMain(os.Args[1:]))
}

func doMain(args []string) int {
	cmd := flag.NewFlagSet("pixdexer", flag.ContinueOnError)
	dir := cmd.String("d", "", "Directory to index")

	err := cmd.Parse(args)
	if err != nil {
		if err != flag.ErrHelp {
			logrus.Errorf("error while parsing command line arguments: %v", err)
		}
		return retConfFailure
	}

	opts := []func(*indexer.Indexer) error{}
	if *dir=="" {
		logrus.Errorf("No directory provided")
		return retConfFailure
	}
	opts = append(opts, indexer.Input(*dir))

	idxer, err := indexer.NewIndexer(opts...)
	if err != nil {
		logrus.Errorf("error while initializing indexer: %v", err)
		return retExecFailure
	}
	defer idxer.Close()

	if err:= idxer.Index() ; err != nil {
		logrus.Errorf("error while indexing: %v", err)
		return retExecFailure
	}

	

	return retOk
}
