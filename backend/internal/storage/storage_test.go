package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withTempImageDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	old := imageDir
	imageDir = dir
	t.Cleanup(func() {
		imageDir = old
	})
	return dir
}

func savedPath(dir, urlPath string) string {
	return filepath.Join(dir, strings.TrimPrefix(urlPath, ImageURLPrefix))
}

func TestSaveImageAllowsJPEG(t *testing.T) {
	dir := withTempImageDir(t)
	path, err := SaveImage(bytes.NewReader([]byte{0xff, 0xd8, 0xff, 0xdb, 0x00, 0x43, 0x00, 0x08, 0x06, 0x06, 0x07, 0xff, 0xd9}))
	if err != nil {
		t.Fatalf("SaveImage returned error: %v", err)
	}
	if !strings.HasSuffix(path, ".jpg") {
		t.Fatalf("path = %q, want .jpg suffix", path)
	}
	if _, err := os.Stat(savedPath(dir, path)); err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
}

func TestSaveImageAllowsPNG(t *testing.T) {
	dir := withTempImageDir(t)
	path, err := SaveImage(bytes.NewReader([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52}))
	if err != nil {
		t.Fatalf("SaveImage returned error: %v", err)
	}
	if !strings.HasSuffix(path, ".png") {
		t.Fatalf("path = %q, want .png suffix", path)
	}
	if _, err := os.Stat(savedPath(dir, path)); err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
}

func TestSaveImageAllowsGIF(t *testing.T) {
	dir := withTempImageDir(t)
	path, err := SaveImage(bytes.NewReader([]byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")))
	if err != nil {
		t.Fatalf("SaveImage returned error: %v", err)
	}
	if !strings.HasSuffix(path, ".gif") {
		t.Fatalf("path = %q, want .gif suffix", path)
	}
	if _, err := os.Stat(savedPath(dir, path)); err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
}

func TestSaveImageRejectsEmpty(t *testing.T) {
	dir := withTempImageDir(t)
	_, err := SaveImage(bytes.NewReader(nil))
	if err == nil || err.Error() != "empty file" {
		t.Fatalf("err = %v, want empty file", err)
	}
	assertDirEmpty(t, dir)
}

func TestSaveImageRejectsUnsupported(t *testing.T) {
	dir := withTempImageDir(t)
	_, err := SaveImage(bytes.NewReader([]byte("plain text is not an image")))
	if err == nil {
		t.Fatal("expected unsupported file type to be rejected")
	}
	assertDirEmpty(t, dir)
}

func TestSaveImageRejectsSpoofedContentType(t *testing.T) {
	dir := withTempImageDir(t)
	_, err := SaveImage(bytes.NewReader([]byte("pretend this has an image/png header elsewhere")))
	if err == nil {
		t.Fatal("expected spoofed content to be rejected")
	}
	assertDirEmpty(t, dir)
}

func TestSaveImageRejectsOversized(t *testing.T) {
	dir := withTempImageDir(t)
	payload := append([]byte{0xff, 0xd8, 0xff, 0xdb}, bytes.Repeat([]byte{0x00}, MaxImageSize+1)...)
	_, err := SaveImage(bytes.NewReader(payload))
	if err == nil || err.Error() != "file too large" {
		t.Fatalf("err = %v, want file too large", err)
	}
	assertDirEmpty(t, dir)
}

func TestDeleteImageMissingFile(t *testing.T) {
	withTempImageDir(t)
	err := DeleteImage(ImageURLPrefix + "missing.jpg")
	if err != nil {
		t.Fatalf("DeleteImage returned error: %v", err)
	}
}

func TestDeleteImageRemovesFile(t *testing.T) {
	dir := withTempImageDir(t)
	path, err := SaveImage(bytes.NewReader([]byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")))
	if err != nil {
		t.Fatalf("SaveImage returned error: %v", err)
	}

	if err := DeleteImage(path); err != nil {
		t.Fatalf("DeleteImage returned error: %v", err)
	}
	_, err = os.Stat(savedPath(dir, path))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat err = %v, want not exist", err)
	}
}

func TestReturnedPathIsBrowserLoadable(t *testing.T) {
	withTempImageDir(t)
	path, err := SaveImage(bytes.NewReader([]byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")))
	if err != nil {
		t.Fatalf("SaveImage returned error: %v", err)
	}
	if !strings.HasPrefix(path, ImageURLPrefix) {
		t.Fatalf("path = %q, want prefix %q", path, ImageURLPrefix)
	}
}

func assertDirEmpty(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		t.Fatalf("ReadDir returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("directory %s has %d entries, want empty", dir, len(entries))
	}
}
