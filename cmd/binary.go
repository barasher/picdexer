package cmd

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/barasher/picdexer/internal/binary"
	"github.com/spf13/cobra"
)

var (
	binCmd = &cobra.Command{
		Use:   "binary",
		Short: "Picdexer : binary utilities",
	}

	binSimuCmd = &cobra.Command{
		Use:   "simulate",
		Short: "Simulate binary push",
		RunE:  simulateBin,
	}

	binPushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push binary push",
		RunE:  pushBin,
	}
)

func init() {
	// simulate
	binSimuCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	binSimuCmd.Flags().StringVarP(&input, "dir", "d", "", "Directory/File containing pictures")
	binSimuCmd.Flags().StringVarP(&output, "out", "o", "", "Output dir")
	binSimuCmd.MarkFlagRequired("conf")
	binSimuCmd.MarkFlagRequired("dir")
	binSimuCmd.MarkFlagRequired("out")
	binCmd.AddCommand(binSimuCmd)

	// push
	binPushCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	binPushCmd.Flags().StringVarP(&input, "dir", "d", "", "Directory/File containing pictures")
	binPushCmd.MarkFlagRequired("conf")
	binPushCmd.MarkFlagRequired("dir")
	binCmd.AddCommand(binPushCmd)

	rootCmd.AddCommand(binCmd)
}

func doBin(push bool) error {
	var c conf.Conf
	var err error
	if confFile != "" {
		if c, err = conf.LoadConf(confFile); err != nil {
			return fmt.Errorf("error while loading configuration (%v): %w", confFile, err)
		}
	}

	s, err := binary.NewStorer(c.Binary, push)
	if err != nil {
		return fmt.Errorf("error while initializing storer: %w", err)
	}

	ctx := context.Background()
	s.StoreFolder(ctx, input, output)

	return nil
}

func simulateBin(cmd *cobra.Command, args []string) error {
	return doBin(false)
}

func pushBin(cmd *cobra.Command, arfs []string) error {
	return doBin(true)
}
