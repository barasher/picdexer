package conf

import (
	"encoding/json"
	"fmt"
	"os"
)

type Conf struct {
	LogLevel string // TODO gérer
	Elasticsearch ElasticsearchConf `json:"elasticsearch"`
}

type ElasticsearchConf struct {
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
