package cmd

import (
	"fmt"
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
	var c Config
	var err error
	if confFile != "" {
		if c, err = LoadConf(confFile); err != nil {
			return fmt.Errorf("error while loading configuration (%v): %w", confFile, err)
		}
	}

	if err := setLoggingLevel(c.LogLevel); err != nil {
		return fmt.Errorf("error while configuring logging level: %w", err)
	}

	s, err := setup.NewSetup(c.Elasticsearch.Url, c.Kibana.Url, c.Binary.Url)
	if err != nil {
		return fmt.Errorf("Setup initialization error: %w", err)
	}

	if err := s.SetupElasticsearch(); err != nil {
		return fmt.Errorf("error while configuring Elasticsearch: %w", err)
	}
	if err := s.SetupKibana(); err != nil {
		return fmt.Errorf("error while configuring Kibana: %w", err)
	}

	return nil
}
