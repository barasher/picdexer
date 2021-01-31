package binary

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const resizedFileIdentifier = "resizedFile"

type resizerInterface interface {
	resize(ctx context.Context, f string, o string) (string, string, error)
	cleanup(ctx context.Context, f string) error
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
	dimensions  string
	fallbackExt []string
}

func (r resizer) resize(ctx context.Context, f string, d string) (string, string, error) {
	outFilename, err := getOutputFilename(f)
	if err != nil {
		return "", "", fmt.Errorf("error while calculating output filename for %v: %w", f, err)
	}
	outPath := filepath.Join(d, outFilename)
	var cmd *exec.Cmd
	if r.hasToFallback(f) {
		args := fmt.Sprintf("exiftool %v -b -previewImage | convert - -size %v %v", f, r.dimensions, outPath)
		fmt.Println(args)
		cmd = exec.Command("bash", "-c", args)
	} else {
		args := []string{f, "-quiet", "-resize", r.dimensions, outPath}
		cmd = exec.Command("convert", args...)
	}
	b, _ := cmd.CombinedOutput()
	if len(b) > 0 {
		return "", "", fmt.Errorf("error on stdout %v: %v", f, string(b))
	}
	return outPath, outFilename, nil
}

func (r resizer) hasToFallback(f string) bool {
	if len(r.fallbackExt) > 0 {
		lf := strings.ToLower(f)
		for _, curExt := range r.fallbackExt {
			if strings.HasSuffix(lf, curExt) {
				return true
			}
		}
	}
	return false
}

func (r resizer) cleanup(ctx context.Context, f string) error {
	return os.Remove(f)
}

func NewResizer(w int, h int, fallbackExtensions []string) resizer {
	r := resizer{
		dimensions: fmt.Sprintf("%vx%v", w, h),
	}
	r.fallbackExt = make([]string, len(fallbackExtensions))
	for i, cur := range fallbackExtensions {
		r.fallbackExt[i] = strings.ToLower(cur)
	}
	return r
}

type nopResizer struct {
}

func (r nopResizer) resize(ctx context.Context, f string, d string) (string, string, error) {
	outFilename, err := getOutputFilename(f)
	if err != nil {
		return "", "", fmt.Errorf("error while calculating output filename for %v: %w", f, err)
	}
	return f, outFilename, nil
}

func (r nopResizer) cleanup(ctx context.Context, f string) error {
	return nil
}

func NewNopResizer() nopResizer {
	return nopResizer{}
}
