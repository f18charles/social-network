package utils

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid/v5"
)

const (
	MaxImageSize = 5 * 1024 * 1024 // 5 MiB
	ImageDir     = "./uploads/images"
)

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
}

func SaveImage(file io.ReadSeeker, directory string) (string, error) {
	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	if n == 0 {
		return "", errors.New("empty file")
	}

	contentType := http.DetectContentType(buffer[:n])

	ext, ok := allowedTypes[contentType]
	if !ok {
		return "", errors.New("unsupported file type")
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	if directory == "" {
		directory = ImageDir
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	filename := id.String() + ext
	fullPath := filepath.Join(directory, filename)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		_ = os.Remove(fullPath)
		return "", err
	}

	return directory + filename, nil
}

func DeleteImage(path string) error {
	if path == "" {
		return nil
	}

	err := os.Remove("." + path)

	if os.IsNotExist(err) {
		return nil
	}

	return err
}
