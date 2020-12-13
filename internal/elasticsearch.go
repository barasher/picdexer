package internal

import "context"

type EsDoc struct {
	Header struct {
		Index string `json:"_index"`
		ID    string `json:"_id"`
	} `json:"index"`
	Document interface{}
}

func BuildEsDocuments(ctx context.Context, inModelChan chan PictureMetadata, outEsDocChan chan EsDoc) error {
	return nil
}

func PushToEs(ctx context.Context, inEsDocChan chan EsDoc) error {
	return nil
}

func PrintEsDocuments(ctx context.Context, inEsDocChan chan EsDoc) error {
	return nil
}
