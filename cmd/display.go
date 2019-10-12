package cmd

import (
	"fmt"
	"os"

	exif "github.com/barasher/go-exiftool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var displayCmd = &cobra.Command{
	Use:   "display",
	Short: "Display file metadata",
	Run:   Display,
}

func init() {
	displayCmd.Flags().StringVarP(&input, "file", "f", "", "File to extract")
	displayCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(displayCmd)
}

func Display(cmd *cobra.Command, args []string) {
	ret, err := doDisplay()
	if err != nil {
		logrus.Errorf("%v", err)
	}
	os.Exit(ret)
}

func doDisplay() (int, error) {
	et, err := exif.NewExiftool()
	if err != nil {
		return retExecFailure, fmt.Errorf("error while initializing metadata extractor: %v", err)
	}
	defer et.Close()

	metas := et.ExtractMetadata(input)
	if len(metas) != 1 {
		return retExecFailure, fmt.Errorf("wrong metadatas count (%v)", len(metas))
	}

	if metas[0].Err != nil {
		return retExecFailure, fmt.Errorf("Error while extracting metadatas: %v", metas[0].Err)
	}

	for k, v := range metas[0].Fields {
		fmt.Printf("%v: %v\n", k, v)
	}

	return retOk, nil
}
