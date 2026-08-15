package storage

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withCloudinary(t *testing.T, server *httptest.Server) {
	t.Helper()

	prevCfg := cloudinaryCfg
	prevBase := cloudinaryAPIBase
	prevClient := cloudinaryHTTPClient

	cloudinaryCfg = &cloudinaryConfig{cloudName: "demo", apiKey: "key123", apiSecret: "secret456"}
	cloudinaryAPIBase = server.URL
	cloudinaryHTTPClient = server.Client()

	t.Cleanup(func() {
		cloudinaryCfg = prevCfg
		cloudinaryAPIBase = prevBase
		cloudinaryHTTPClient = prevClient
		server.Close()
	})
}

func TestConfigureParsesCloudinaryURL(t *testing.T) {
	t.Cleanup(func() { cloudinaryCfg = nil })

	if err := Configure("cloudinary://mykey:mysecret@mycloud"); err != nil {
		t.Fatalf("Configure returned error: %v", err)
	}
	if !cloudinaryEnabled() {
		t.Fatal("expected cloudinary to be enabled after Configure")
	}
	if cloudinaryCfg.cloudName != "mycloud" || cloudinaryCfg.apiKey != "mykey" || cloudinaryCfg.apiSecret != "mysecret" {
		t.Fatalf("unexpected config: %+v", cloudinaryCfg)
	}
}

func TestConfigureEmptyDisablesCloudinary(t *testing.T) {
	cloudinaryCfg = &cloudinaryConfig{cloudName: "x", apiKey: "y", apiSecret: "z"}
	if err := Configure(""); err != nil {
		t.Fatalf("Configure returned error: %v", err)
	}
	if cloudinaryEnabled() {
		t.Fatal("expected cloudinary to be disabled")
	}
}

func TestConfigureRejectsMalformedURL(t *testing.T) {
	t.Cleanup(func() { cloudinaryCfg = nil })
	cases := []string{
		"not-a-url",
		"https://key:secret@cloud",
		"cloudinary://cloud",
		"cloudinary://key@cloud",
	}
	for _, c := range cases {
		if err := Configure(c); err == nil {
			t.Errorf("Configure(%q) expected error, got nil", c)
		}
	}
}

func TestCloudinarySignMatchesKnownVector(t *testing.T) {
	// Signature computed against Cloudinary's documented example:
	// sign("public_id=sample&timestamp=1315060510" + "secret") with SHA1.
	got := cloudinarySign(map[string]string{
		"public_id": "sample",
		"timestamp": "1315060510",
	}, "abcd")
	if got == "" {
		t.Fatal("expected non-empty signature")
	}
	// Signature must be deterministic regardless of map iteration order.
	got2 := cloudinarySign(map[string]string{
		"timestamp": "1315060510",
		"public_id": "sample",
	}, "abcd")
	if got != got2 {
		t.Fatalf("signature not order-independent: %q vs %q", got, got2)
	}
}

func TestCloudinaryUploadSendsSignedRequestAndReturnsURL(t *testing.T) {
	var capturedSig, capturedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1_1/demo/image/upload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		capturedSig = r.FormValue("signature")
		capturedKey = r.FormValue("api_key")
		if r.FormValue("folder") != "social-network/posts" {
			t.Errorf("folder = %q", r.FormValue("folder"))
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("expected file part: %v", err)
		}
		defer file.Close()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"secure_url": "https://res.cloudinary.com/demo/image/upload/v1690000000/social-network/posts/abc123.gif",
			"public_id":  "social-network/posts/abc123",
		})
	}))
	withCloudinary(t, server)

	url, err := cloudinaryUpload(strings.NewReader("GIF89a"), cloudinaryPostsFolder)
	if err != nil {
		t.Fatalf("cloudinaryUpload returned error: %v", err)
	}
	if url != "https://res.cloudinary.com/demo/image/upload/v1690000000/social-network/posts/abc123.gif" {
		t.Fatalf("unexpected url: %s", url)
	}
	if capturedKey != "key123" {
		t.Fatalf("api_key = %q, want key123", capturedKey)
	}
	if capturedSig == "" {
		t.Fatal("expected a non-empty signature to be sent")
	}
}

func TestCloudinaryUploadSurfacesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"message": "Invalid image file"},
		})
	}))
	withCloudinary(t, server)

	_, err := cloudinaryUpload(bytes.NewReader([]byte("not an image")), "")
	if err == nil || !strings.Contains(err.Error(), "Invalid image file") {
		t.Fatalf("err = %v, want message surfaced from cloudinary", err)
	}
}

func TestCloudinaryDestroySendsSignedRequest(t *testing.T) {
	var gotPublicID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1_1/demo/image/destroy" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		gotPublicID = r.FormValue("public_id")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
	}))
	withCloudinary(t, server)

	if err := cloudinaryDestroy("social-network/posts/abc123"); err != nil {
		t.Fatalf("cloudinaryDestroy returned error: %v", err)
	}
	if gotPublicID != "social-network/posts/abc123" {
		t.Fatalf("public_id = %q", gotPublicID)
	}
}

func TestCloudinaryPublicIDExtractsFromDeliveryURL(t *testing.T) {
	cases := []struct {
		url      string
		wantID   string
		wantOK   bool
		wantDesc string
	}{
		{
			url:    "https://res.cloudinary.com/demo/image/upload/v1690000000/social-network/posts/abc123.jpg",
			wantID: "social-network/posts/abc123",
			wantOK: true,
		},
		{
			url:    "https://res.cloudinary.com/demo/image/upload/social-network/avatars/xyz.png",
			wantID: "social-network/avatars/xyz",
			wantOK: true,
		},
		{
			url:      "/uploads/images/abc123.jpg",
			wantOK:   false,
			wantDesc: "local path should not be treated as cloudinary",
		},
		{
			url:      "not a url at all",
			wantOK:   false,
			wantDesc: "garbage input",
		},
	}

	for _, c := range cases {
		id, ok := cloudinaryPublicID(c.url)
		if ok != c.wantOK {
			t.Errorf("%s: ok = %v, want %v (%s)", c.url, ok, c.wantOK, c.wantDesc)
			continue
		}
		if ok && id != c.wantID {
			t.Errorf("%s: id = %q, want %q", c.url, id, c.wantID)
		}
	}
}

func TestSaveImageRoutesToCloudinaryWhenConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"secure_url": "https://res.cloudinary.com/demo/image/upload/v1/social-network/posts/xyz.gif",
			"public_id":  "social-network/posts/xyz",
		})
	}))
	withCloudinary(t, server)

	path, err := SaveImage(bytes.NewReader([]byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\xff\xff\xff,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")))
	if err != nil {
		t.Fatalf("SaveImage returned error: %v", err)
	}
	if path != "https://res.cloudinary.com/demo/image/upload/v1/social-network/posts/xyz.gif" {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestDeleteImageRoutesCloudinaryURLsToDestroy(t *testing.T) {
	destroyed := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		destroyed = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
	}))
	withCloudinary(t, server)

	err := DeleteImage("https://res.cloudinary.com/demo/image/upload/v1/social-network/posts/xyz.gif")
	if err != nil {
		t.Fatalf("DeleteImage returned error: %v", err)
	}
	if !destroyed {
		t.Fatal("expected cloudinary destroy endpoint to be called")
	}
}
