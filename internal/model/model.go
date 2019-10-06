package model

import (
	"time"
)

type Model struct {
	Aperture     float32   // Aperture
	ShutterSpeed string    // ShutterSpeed
	Keywords     []string  // Keywords
	CameraModel  string    // Model
	LensModel    string    // LensModel
	MimeType     string    // MIMEType
	Height       uint32    // ImageHeight
	Width        uint32    // ImageWidth
	FileSize     uint32    //
	CaptureDate  time.Time // CaptureDate
	GPS          string    // GPSPosition "41.12,-71.34"
}
