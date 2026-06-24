package storage

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid/v5"
)

const (
	MaxImageSize   = 5 * 1024 * 1024
	ImageURLPrefix = "/uploads/images/"
	ImageDir       = "./uploads/images"
)

var (
	allowedTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
	}

	imageDir = ImageDir
)

func SaveImage(file io.ReadSeeker) (string, error) {
	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if errors.Is(err, io.EOF) && n == 0 {
		return "", errors.New("empty file")
	}
	if err != nil && !errors.Is(err, io.EOF) {
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

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		return "", err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	filename := id.String() + ext
	fullPath := filepath.Join(imageDir, filename)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(file, MaxImageSize+1))
	if err != nil {
		_ = os.Remove(fullPath)
		return "", err
	}
	if written > MaxImageSize {
		_ = os.Remove(fullPath)
		return "", errors.New("file too large")
	}

	return ImageURLPrefix + filename, nil
}

func DeleteImage(urlPath string) error {
	if urlPath == "" {
		return nil
	}

	cleanURL := path.Clean("/" + strings.TrimPrefix(urlPath, "/"))
	if !strings.HasPrefix(cleanURL, ImageURLPrefix) {
		return fmt.Errorf("image path outside storage directory: %s", urlPath)
	}

	rel := strings.TrimPrefix(cleanURL, ImageURLPrefix)
	if rel == "" || rel == "." || strings.Contains(rel, "..") {
		return fmt.Errorf("invalid image path: %s", urlPath)
	}

	fullPath := filepath.Join(imageDir, filepath.FromSlash(rel))
	baseDir, err := filepath.Abs(imageDir)
	if err != nil {
		return err
	}
	target, err := filepath.Abs(fullPath)
	if err != nil {
		return err
	}
	if target != baseDir && !strings.HasPrefix(target, baseDir+string(os.PathSeparator)) {
		return fmt.Errorf("image path outside storage directory: %s", urlPath)
	}

	err = os.Remove(target)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
