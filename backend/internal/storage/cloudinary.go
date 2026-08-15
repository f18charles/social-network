package storage

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// cloudinaryConfig holds the parsed credentials for a Cloudinary account.
type cloudinaryConfig struct {
	cloudName string
	apiKey    string
	apiSecret string
}

var (
	// cloudinaryCfg is nil when Cloudinary is not configured, in which case
	// SaveImage/SaveAvatar/DeleteImage fall back to local disk storage.
	cloudinaryCfg *cloudinaryConfig

	// cloudinaryAvatarsFolder and cloudinaryPostsFolder namespace uploads
	// inside the Cloudinary account so they're easy to find and manage.
	cloudinaryAvatarsFolder = "social-network/avatars"
	cloudinaryPostsFolder   = "social-network/posts"

	// cloudinaryHTTPClient is overridable in tests.
	cloudinaryHTTPClient = &http.Client{Timeout: 30 * time.Second}

	// cloudinaryAPIBase is overridable in tests to point at a mock server.
	cloudinaryAPIBase = "https://api.cloudinary.com"
)

// versionSegment matches a Cloudinary delivery URL version component, e.g. "v1690000000".
var versionSegment = regexp.MustCompile(`^v[0-9]+$`)

// Configure enables Cloudinary-backed storage using the given account URL,
// which must be in the form cloudinary://<api_key>:<api_secret>@<cloud_name>.
// Passing an empty string disables Cloudinary and restores local disk
// storage, which is the default and is suitable for local development.
func Configure(cloudinaryURL string) error {
	cloudinaryURL = strings.TrimSpace(cloudinaryURL)
	if cloudinaryURL == "" {
		cloudinaryCfg = nil
		return nil
	}

	parsed, err := url.Parse(cloudinaryURL)
	if err != nil {
		return fmt.Errorf("invalid CLOUDINARY_URL: %w", err)
	}
	if parsed.Scheme != "cloudinary" || parsed.Host == "" || parsed.User == nil {
		return errors.New("invalid CLOUDINARY_URL: expected cloudinary://<api_key>:<api_secret>@<cloud_name>")
	}

	apiKey := parsed.User.Username()
	apiSecret, hasSecret := parsed.User.Password()
	if apiKey == "" || !hasSecret || apiSecret == "" {
		return errors.New("invalid CLOUDINARY_URL: missing API key or secret")
	}

	cloudinaryCfg = &cloudinaryConfig{
		cloudName: parsed.Host,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
	return nil
}

// cloudinaryEnabled reports whether uploads should be routed to Cloudinary.
func cloudinaryEnabled() bool {
	return cloudinaryCfg != nil
}

type cloudinaryUploadResponse struct {
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// cloudinaryUpload signs and sends the given file contents to Cloudinary's
// upload API, namespacing it under folder. It returns the HTTPS delivery URL
// Cloudinary assigns to the asset.
func cloudinaryUpload(file io.Reader, folder string) (string, error) {
	if !cloudinaryEnabled() {
		return "", errors.New("cloudinary is not configured")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{"timestamp": timestamp}
	if folder != "" {
		params["folder"] = folder
	}
	signature := cloudinarySign(params, cloudinaryCfg.apiSecret)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range params {
		if err := writer.WriteField(key, value); err != nil {
			return "", err
		}
	}
	if err := writer.WriteField("api_key", cloudinaryCfg.apiKey); err != nil {
		return "", err
	}
	if err := writer.WriteField("signature", signature); err != nil {
		return "", err
	}

	part, err := writer.CreateFormFile("file", "upload")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/v1_1/%s/image/upload", cloudinaryAPIBase, cloudinaryCfg.cloudName)
	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := cloudinaryHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload request failed: %w", err)
	}
	defer resp.Body.Close()

	var result cloudinaryUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("cloudinary returned an unreadable response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if result.Error != nil && result.Error.Message != "" {
			return "", fmt.Errorf("cloudinary upload failed: %s", result.Error.Message)
		}
		return "", fmt.Errorf("cloudinary upload failed with status %d", resp.StatusCode)
	}
	if result.SecureURL == "" {
		return "", errors.New("cloudinary upload response missing secure_url")
	}

	return result.SecureURL, nil
}

// cloudinaryDestroy deletes the asset identified by publicID from Cloudinary.
func cloudinaryDestroy(publicID string) error {
	if !cloudinaryEnabled() {
		return errors.New("cloudinary is not configured")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"public_id": publicID,
		"timestamp": timestamp,
	}
	signature := cloudinarySign(params, cloudinaryCfg.apiSecret)

	form := url.Values{}
	form.Set("public_id", publicID)
	form.Set("timestamp", timestamp)
	form.Set("api_key", cloudinaryCfg.apiKey)
	form.Set("signature", signature)

	endpoint := fmt.Sprintf("%s/v1_1/%s/image/destroy", cloudinaryAPIBase, cloudinaryCfg.cloudName)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := cloudinaryHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloudinary destroy request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloudinary destroy failed with status %d", resp.StatusCode)
	}
	return nil
}

// cloudinarySign implements Cloudinary's request-signing algorithm: params
// are sorted by key, joined as key=value pairs with "&", the api secret is
// appended, and the result is SHA1-hashed.
func cloudinarySign(params map[string]string, apiSecret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+params[key])
	}

	toSign := strings.Join(pairs, "&") + apiSecret
	sum := sha1.Sum([]byte(toSign))
	return hex.EncodeToString(sum[:])
}

// cloudinaryPublicID extracts the public ID from a Cloudinary delivery URL,
// e.g. https://res.cloudinary.com/demo/image/upload/v169/social-network/posts/abc123.jpg
// yields "social-network/posts/abc123". It reports false if rawURL isn't a
// recognizable Cloudinary delivery URL.
func cloudinaryPublicID(rawURL string) (string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil || !strings.Contains(parsed.Host, "cloudinary.com") {
		return "", false
	}

	const marker = "/upload/"
	idx := strings.Index(parsed.Path, marker)
	if idx == -1 {
		return "", false
	}

	remainder := strings.TrimPrefix(parsed.Path[idx+len(marker):], "/")
	segments := strings.Split(remainder, "/")
	if len(segments) == 0 {
		return "", false
	}
	if versionSegment.MatchString(segments[0]) {
		segments = segments[1:]
	}
	if len(segments) == 0 {
		return "", false
	}

	joined := strings.Join(segments, "/")
	if ext := strings.LastIndex(joined, "."); ext != -1 {
		joined = joined[:ext]
	}
	if joined == "" {
		return "", false
	}
	return joined, true
}
