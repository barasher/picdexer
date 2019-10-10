package model

import (
	"time"
)

type Model struct {
	Aperture     float32
	ShutterSpeed string
	Keywords     []string
	CameraModel  string
	LensModel    string
	MimeType     string
	Height       uint32
	Width        uint32
	FileSize     uint32
	CaptureDate  time.Time `json:"-"`
	GPS          string
}
