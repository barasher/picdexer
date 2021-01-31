package cmd

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/internal/common"
	"github.com/barasher/picdexer/internal/filewatcher"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	dzCmd = &cobra.Command{
		Use:   "dropzone",
		Short: "Picdexer : Dropzone",
		RunE:  dropzone,
	}
)

func init() {
	// full
	dzCmd.Flags().StringVarP(&confFile, "conf", "c", "", "Picdexer configuration file")
	dzCmd.Flags().StringVarP(&importID, "impId", "i", "", "Import identifier")

	dzCmd.MarkFlagRequired("conf")
	rootCmd.AddCommand(dzCmd)
}

func dropzone(cmd *cobra.Command, args []string) error {
	return dropzone2(confFile, importID, Run)
}

func dropzone2(confFile string, importId string, runFct func(context.Context, Config, []string) error) error {
	ctx := common.NewContext(importId)
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
	return doDropzone(ctx, c, runFct)
}

func doDropzone(ctx context.Context, c Config, runFct func(context.Context, Config, []string) error) error {
	fw := filewatcher.NewFileWatcher(c.Dropzone.Root)
	fw.Watch()

	period, err := time.ParseDuration(c.Dropzone.Period)
	if err != nil {
		return fmt.Errorf("error while parsing watching duration (%s): %w", c.Dropzone.Period, err)
	}
	timer := time.NewTimer(period)

	err = process(ctx, fw, c, runFct)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			err = process(ctx, fw, c, runFct)
			if err != nil {
				return err
			}
			timer.Reset(period)
		}
	}

	return nil
}

func process(ctx context.Context, fw *filewatcher.FileWatcher, c Config, runFct func(context.Context, Config, []string) error) error {
	watched, err := fw.Watch()
	if err != nil {
		return fmt.Errorf("error while watching folder: %w", err)
	}

	// process
	inputs := make([]string, len(watched))
	for i, c := range watched {
		inputs[i] = c.Path
	}
	err = runFct(ctx, c, inputs)
	if err != nil {
		return fmt.Errorf("error while running: %w", err)
	}

	// delete
	for _, curInput := range inputs {
		os.Remove(curInput)
	}
	return nil
}
