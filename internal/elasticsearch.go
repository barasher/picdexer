package internal

import "context"

const (

	defaultBulkSize = 30
)

type EsDoc struct {
	Header struct {
		Index string `json:"_index"`
		ID    string `json:"_id"`
	} `json:"index"`
	Document interface{}
}

type EsPusher struct {
	conf EsPusherConf
}

type EsPusherConf struct {
	BulkSize int    `json:"bulkSize"`
}

func NewEsPusher(conf EsPusherConf) (*EsPusher, error) {
	p := EsPusher{conf: conf}
	return &p, nil
}

func (p *EsPusher) bulkSize() int {
	n := p.conf.BulkSize
	if n < 1 {
		n = defaultBulkSize
	}
	return n
}

func (*EsPusher) Push(ctx context.Context, inEsDocChan chan EsDoc) error {
	return nil
}

func (*EsPusher) Print(ctx context.Context, inEsDocChan chan EsDoc) error {
	return nil
}
