package model

type Model struct {
	Aperture     *float64 `json:",omitempty"`
	ShutterSpeed *string  `json:",omitempty"`
	Keywords     []string `json:",omitempty"`
	CameraModel  *string  `json:",omitempty"`
	LensModel    *string  `json:",omitempty"`
	MimeType     *string  `json:",omitempty"`
	Height       *uint64  `json:",omitempty"`
	Width        *uint64  `json:",omitempty"`
	FileSize     uint64
	Date         *uint64 `json:",omitempty"`
	GPS          *string `json:",omitempty"`
	FileName     string
	Folder       string
}
