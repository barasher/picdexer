package internal

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const resizedFileIdentifier = "resizedFile"

type resizerInterface interface {
	resize(ctx context.Context, f string, o string) (string, string, error)
}

func getOutputFilename(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("error while opening file %v: %w", f, err)
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("error while calculating md5 for %v: %w", file, err)
	}
	o := fmt.Sprintf("%s_%s", hex.EncodeToString(h.Sum(nil)), filepath.Base(file))
	return o, nil
}

type resizer struct {
	dimensions string
}


func (r resizer) resize(ctx context.Context, f string, d string) (string, string, error) {
	outFilename, err := getOutputFilename(f)
	if err != nil {
		return "", "", fmt.Errorf("error while calculating output filename for %v: %w", f, err)
	}
	outPath := filepath.Join(d, outFilename)
	args := []string{f, "-quiet", "-resize", r.dimensions, outPath}
	cmd := exec.Command("convert", args...)
	b, _ := cmd.CombinedOutput()
	if len(b) > 0 {
		return "", "", fmt.Errorf("error on stdout %v: %v", f, string(b))
	}
	return outPath, outFilename, nil
}

func NewResizer(w, h int) resizer {
	return resizer{dimensions: fmt.Sprintf("%vx%v", w, h)}
}

type nopResizer struct {
}

func (r nopResizer) resize(ctx context.Context, f string, d string) (string, string,  error) {
	outFilename, err := getOutputFilename(f)
	if err != nil {
		return "", "", fmt.Errorf("error while calculating output filename for %v: %w", f, err)
	}
	return f, outFilename, nil
}

func NewNopResizer() nopResizer {
	return nopResizer{}
}

