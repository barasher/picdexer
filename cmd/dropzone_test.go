package cmd

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestFailOnWrongLoggingLevel(t *testing.T) {
	assert.NotNil(t, dropzone2("../testdata/conf/picdexer_wrongLoggingLevel.json", "", simulateRun(true)))
}

func TestFailOnConfLoad(t *testing.T) {
	assert.NotNil(t, dropzone2("nonExistingFile", "", simulateRun(true)))
}

func TestFailOnRun(t *testing.T) {
	d, err := ioutil.TempDir("/tmp/", "TestFailOnRun")
	assert.Nil(t, err)
	assert.Nil(t, copy("../testdata/picture.jpg", d+"/picture.jpg"))
	//defer os.RemoveAll(d)
	c := Config{Dropzone: DropzoneConf{
		Root:   d,
		Period: "10ms",
	}}
	assert.NotNil(t, doDropzone(context.Background(), c, simulateRun(false)))
}

func TestFailOnNonExistingRootFolder(t *testing.T) {
	c := Config{Dropzone: DropzoneConf{
		Root:   "/tmp123456789",
		Period: "10ms",
	}}
	assert.NotNil(t, doDropzone(context.Background(), c, simulateRun(true)))
}

func TestDoDropzone_UnparsablePeriod(t *testing.T) {
	c := Config{Dropzone: DropzoneConf{
		Root:   "/tmp",
		Period: "blubla",
	}}
	assert.NotNil(t, doDropzone(context.Background(), c, simulateRun(true)))
}

func TestDropzone_nominal(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "TestWatch")
	assert.Nil(t, err)
	t.Logf("tmpDir: %v", tmpDir)
	defer os.RemoveAll(tmpDir)
	dest := tmpDir + "/picture.jpg"
	copy("../testdata/picture.jpg", dest)

	ctx, cancel := context.WithCancel(context.Background())
	conf := Config{
		Dropzone: DropzoneConf{
			Root:   tmpDir,
			Period: "10ms",
		},
	}

	watched := []string{}
	fct := func(ctx2 context.Context, conf2 Config, inputs []string) error {
		watched = append(watched, inputs...)
		for _, cur := range inputs {
			os.Remove(cur)
		}
		return nil
	}
	go doDropzone(ctx, conf, fct)

	time.Sleep(100 * time.Millisecond)
	cancel()

	assert.ElementsMatch(t, []string{dest}, watched)
}

func copy(src, dest string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, input, 0644)
	if err != nil {
		return err
	}

	return nil
}
