package model

type Model struct {
	Aperture     *float64
	ShutterSpeed *string
	Keywords     []string
	CameraModel  *string
	LensModel    *string
	MimeType     *string
	Height       *uint64
	Width        *uint64
	FileSize     uint64
	Date         *uint64
	GPS          *string
	FileName     string
	Folder       string
}
