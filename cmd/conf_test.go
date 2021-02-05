package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadConf_Nominal(t *testing.T)  {
	c, err := LoadConf("../testdata/conf/picdexer_nominal.json")
	assert.Nil(t, err)
	assert.Equal(t, "debug", c.LogLevel)
	// elasticsearch
	assert.Equal(t, "http://localhost:9200", c.Elasticsearch.Url)
	assert.Equal(t, 10, c.Elasticsearch.ThreadCount)
	assert.Equal(t, 200, c.Elasticsearch.BulkSize)
	// binary
	assert.Equal(t, "http://localhost:8080", c.Binary.Url)
	assert.Equal(t, 480, c.Binary.Height)
	assert.Equal(t, 640, c.Binary.Width)
	assert.Equal(t, 11, c.Binary.ThreadCount)
	assert.Equal(t, "/tmp", c.Binary.WorkingDir)
	assert.Equal(t, []string{"ext1", "ext2"}, c.Binary.UsePreviewForExtensions)
	// kibana
	assert.Equal(t, "http://localhost:5601", c.Kibana.Url)
	// dropzone
	assert.Equal(t, "/tmp2", c.Dropzone.Root)
	assert.Equal(t, "20s", c.Dropzone.Period)
}

func TestLoadConf_NonExistingFile(t *testing.T) {
	_, err := LoadConf("nonExistingFile")
	assert.NotNil(t, err)
}