package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const (
	retOk          int = 0
	retConfFailure int = 1
	retExecFailure int = 2
)

var (
	rootCmd = &cobra.Command{
		Use:   "picdexer",
		Short: "Picture indexer",
	}
)

var input string
var esUrl string
var importID string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
