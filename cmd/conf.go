package cmd

type Config struct {
	LogLevel      string            `json:"loggingLevel"`
	Elasticsearch ElasticsearchConf `json:"elasticsearch"`
	Binary        BinaryConf        `json:"binary"`
}

type ElasticsearchConf struct {
	Url         string `json:"url"`
	ThreadCount int    `json:"threadCount"`
	BulkSize    int    `json:"bulkSize"`
}

type BinaryConf struct {
	Url         string `json:"url"`
	Height      int    `json:"height"`
	Width       int    `json:"width"`
	ThreadCount int    `json:"threadCount"`
	WorkingDir string `json:"workingDir"`
}
