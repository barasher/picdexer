package indexer

import (
	"time"
)

type Model struct {
	FileName     string
	Folder       string
	ImportID     string
	FileSize     uint64
	ISO          *uint64    `json:",omitempty"`
	Aperture     *float64   `json:",omitempty"`
	ShutterSpeed *string    `json:",omitempty"`
	Keywords     []string   `json:",omitempty"`
	CameraModel  *string    `json:",omitempty"`
	LensModel    *string    `json:",omitempty"`
	MimeType     *string    `json:",omitempty"`
	Height       *uint64    `json:",omitempty"`
	Width        *uint64    `json:",omitempty"`
	Date         *uint64    `json:",omitempty"`
	ParsedDate   *time.Time `json:"-"`
	GPS          *string    `json:",omitempty"`
}
