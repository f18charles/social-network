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
	MaxImageSize    = 5 * 1024 * 1024
	ImageURLPrefix  = "/uploads/images/"
	ImageDir        = "./uploads/images"
	AvatarURLPrefix = "/uploads/avatars/"
	AvatarDir       = "./uploads/avatars"
)

var (
	allowedTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
	}

	imageDir  = ImageDir
	avatarDir = AvatarDir
)

// SaveImage stores a post/comment image or GIF. When Cloudinary is
// configured (see Configure), the file is uploaded there and a Cloudinary
// delivery URL is returned; otherwise it is written to local disk.
func SaveImage(file io.ReadSeeker) (string, error) {
	return saveImage(file, imageDir, ImageURLPrefix, cloudinaryPostsFolder)
}

// SaveAvatar stores a user avatar image, following the same Cloudinary or
// local-disk routing as SaveImage.
func SaveAvatar(file io.ReadSeeker) (string, error) {
	return saveImage(file, avatarDir, AvatarURLPrefix, cloudinaryAvatarsFolder)
}

func saveImage(file io.ReadSeeker, directory, urlPrefix, cloudFolder string) (string, error) {
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

	// Determine the total size up front (rather than truncating mid-stream)
	// so both the Cloudinary and local-disk paths reject oversized uploads
	// identically, before any bytes are sent anywhere.
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return "", err
	}
	if size > MaxImageSize {
		return "", errors.New("file too large")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	if cloudinaryEnabled() {
		return cloudinaryUpload(file, cloudFolder)
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

	if _, err := io.Copy(dst, file); err != nil {
		_ = os.Remove(fullPath)
		return "", err
	}

	return urlPrefix + filename, nil
}

// DeleteImage removes a previously saved image, whether it lives on
// Cloudinary or on local disk. It infers which store owns urlPath from its
// shape, so callers don't need to track where an image was saved.
func DeleteImage(urlPath string) error {
	if urlPath == "" {
		return nil
	}

	if publicID, ok := cloudinaryPublicID(urlPath); ok {
		if !cloudinaryEnabled() {
			// The URL is a Cloudinary asset but this process has no
			// Cloudinary credentials configured (e.g. they were removed).
			// Nothing more we can do locally.
			return nil
		}
		return cloudinaryDestroy(publicID)
	}

	cleanURL := path.Clean("/" + strings.TrimPrefix(urlPath, "/"))
	store, ok := storeForURL(cleanURL)
	if !ok {
		return fmt.Errorf("image path outside storage directory: %s", urlPath)
	}

	rel := strings.TrimPrefix(cleanURL, store.urlPrefix)
	if rel == "" || rel == "." || strings.Contains(rel, "..") {
		return fmt.Errorf("invalid image path: %s", urlPath)
	}

	fullPath := filepath.Join(store.directory, filepath.FromSlash(rel))
	baseDir, err := filepath.Abs(store.directory)
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

type imageStore struct {
	urlPrefix string
	directory string
}

func storeForURL(urlPath string) (imageStore, bool) {
	stores := []imageStore{
		{urlPrefix: ImageURLPrefix, directory: imageDir},
		{urlPrefix: AvatarURLPrefix, directory: avatarDir},
	}
	for _, store := range stores {
		if strings.HasPrefix(urlPath, store.urlPrefix) {
			return store, true
		}
	}
	return imageStore{}, false
}
