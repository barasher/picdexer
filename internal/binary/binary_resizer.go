package binary

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const resizedFileIdentifier = "resizedFile"

type resizerInterface interface {
	resize(ctx context.Context, from string, to string) error
	cleanup(ctx context.Context, f string) error
}

type resizer struct {
	dimensions  string
	fallbackExt []string
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

func (r resizer) resize(ctx context.Context, from string, to string) error {
	var cmd *exec.Cmd
	if r.hasToFallback(from) {
		args := fmt.Sprintf("exiftool %v -b -previewImage | convert - -size %v %v", from, r.dimensions, to)
		cmd = exec.Command("bash", "-c", args)
	} else {
		args := []string{from, "-quiet", "-resize", r.dimensions, to}
		cmd = exec.Command("convert", args...)
	}
	b, _ := cmd.CombinedOutput()
	if len(b) > 0 {
		return fmt.Errorf("error on stdout %v: %v", from, string(b))
	}
	return nil
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

func (r nopResizer) resize(ctx context.Context, from string, to string) error {
	return nil
}

func (r nopResizer) cleanup(ctx context.Context, f string) error {
	return nil
}

func NewNopResizer() nopResizer {
	return nopResizer{}
}
