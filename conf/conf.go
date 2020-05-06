package conf

import (
	"encoding/json"
	"fmt"
	"os"
)

type Conf struct {
	LogLevel      string            `json:"loggingLevel"`
	Elasticsearch ElasticsearchConf `json:"elasticsearch"`
	Binary        BinaryConf        `json:"binary"`
	Kibana        KibanaConf        `json:"kibana"`
}

type ElasticsearchConf struct {
	Url                   string `json:"url"`
	ExtractionThreadCount int    `json:"extractionThreadCount"`
	ToExtractChannelSize  int    `json:"toExtractChannelSize"`
}

type BinaryConf struct {
	Url                 string `json:"url"`
	Height              uint   `json:"height"`
	Width               uint   `json:"width"`
	ResizingThreadCount int    `json:"resizingThreadCount"`
	ToResizeChannelSize int    `json:"toResizeChannelSize"`
}

type KibanaConf struct {
	Url string `json:"url"`
}

func LoadConf(f string) (Conf, error) {
	conf := Conf{}
	confReader, err := os.Open(f)
	if err != nil {
		return conf, fmt.Errorf("Error while opening configuration file %v: %w", f, err)
	}
	defer confReader.Close()
	err = json.NewDecoder(confReader).Decode(&conf)
	if err != nil {
		return conf, fmt.Errorf("Error while unmarshaling configuration file %v: %w", f, err)
	}
	return conf, nil
}
