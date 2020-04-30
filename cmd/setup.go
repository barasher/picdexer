package cmd

import (
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/barasher/picdexer/internal/setup"
	"github.com/spf13/cobra"
)

var (
	setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Picdexer : setup components",
		RunE:  configure,
	}
)

func init() {
	setupCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	setupCmd.MarkFlagRequired("conf")
	rootCmd.AddCommand(setupCmd)
}

func configure(cmd *cobra.Command, args []string) error {
	var c conf.Conf
	var err error
	if confFile != "" {
		if c, err = conf.LoadConf(confFile); err != nil {
			return fmt.Errorf("error while loading configuration (%v): %w", confFile, err)
		}
	}

	s, err := setup.NewSetup(c)
	if err != nil {
		return fmt.Errorf("Setup initialization error: %w", err)
	}

	if err := s.SetupElasticsearch() ; err != nil {
		return fmt.Errorf("error while configuring Elasticsearch: %w", err)
	}

	return nil
}