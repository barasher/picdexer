package cmd

import (
	"github.com/rs/zerolog/log"
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
		Short: "Picture metadata",
	}
	input    []string
	output   string
	importID string
	confFile string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Msgf("%v", err)
		os.Exit(1)
	}
}
