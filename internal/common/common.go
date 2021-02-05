package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	LogFileIdentifier   = "file"
	imageMimeTypePrefix = "image/"
	jpegMimeType        = "image/jpeg"
	jpegRefExtension    = ".jpg"
)

func getMimeType(path string) (string, error) {
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		return "", fmt.Errorf("error while getting mime-type for %v: %w", path, err)
	}
	return mime.String(), nil
}

func hash(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("error while opening %v to get hashed: %w", file, err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("error while hashing %v: %w", file, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func CategorizePicture(path string) (bool, string, error) {
	mime, err := getMimeType(path)
	if err != nil {
		return false, "", fmt.Errorf("error while getting mime-type for %v: %w", path, err)
	}
	isPicture := strings.HasPrefix(mime, imageMimeTypePrefix)
	if !isPicture {
		return false, "", nil
	}

	f := filepath.Base(path)
	if mime != jpegMimeType {
		f = f + jpegRefExtension
	}
	h, err := hash(path)
	if err != nil {
		return isPicture, "", fmt.Errorf("error while hashing %v: %w", path, err)
	}
	return isPicture, h + "_" + f, nil
}
