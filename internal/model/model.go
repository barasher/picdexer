package model

type Model struct {
	FileName     string
	Folder       string
	FileSize     uint64
	ID           string
	ISO          *uint64  `json:",omitempty"`
	Aperture     *float64 `json:",omitempty"`
	ShutterSpeed *string  `json:",omitempty"`
	Keywords     []string `json:",omitempty"`
	CameraModel  *string  `json:",omitempty"`
	LensModel    *string  `json:",omitempty"`
	MimeType     *string  `json:",omitempty"`
	Height       *uint64  `json:",omitempty"`
	Width        *uint64  `json:",omitempty"`
	Date         *uint64  `json:",omitempty"`
	GPS          *string  `json:",omitempty"`
}
