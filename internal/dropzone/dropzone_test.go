package dropzone

import (
	"context"
	"fmt"
	"github.com/barasher/picdexer/conf"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

func copy(t *testing.T, src string, dest string) {
	assert.Nil(t, exec.Command("cp", src, dest).Run())
}

func mkdir(t *testing.T, folder string) {
	assert.Nil(t, exec.Command("mkdir", "-p", folder).Run())
}

func TestWatcher(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "dropzoneTest")
	assert.Nil(t, err)
	t.Logf("tmpDir: %s", tmpDir)
	defer os.RemoveAll(tmpDir)

	c := conf.DropzoneConf{
		Root: tmpDir,
	}
	ctx := context.Background()
	w, err := NewWatcher(ctx, c)
	assert.Nil(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var files []string
	go func() {
		for f := range w.FileChan {
			fmt.Printf("%v\n", f)
			files = append(files, f)
		}
		wg.Done()
	}()

	f1 := filepath.Join(tmpDir, "p1.jpg")
	f2 := filepath.Join(tmpDir, "p2.jpg")
	f3 := filepath.Join(tmpDir, "folder", "p2.jpg")
	copy(t, "../../testdata/picture.jpg", f1)
	copy(t, "../../testdata/picture.jpg", f2)
	mkdir(t, filepath.Join(tmpDir, "folder"))
	copy(t, "../../testdata/picture.jpg", f3)
	w.Stop()

	wg.Wait()
	assert.Equal(t, []string{f1, f2, f3}, files)
}
