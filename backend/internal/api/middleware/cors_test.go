package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/config"
)

func TestCorsMiddlewareAllowsCredentialedPatchPreflight(t *testing.T) {
	nextCalled := false
	config.App.AllowedOrigin = "http://localhost:5173"
	handler := CorsMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))

	request := httptest.NewRequest(http.MethodOptions, "/api/posts/123", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}
	if nextCalled {
		t.Fatal("preflight request reached the next handler")
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want true", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodPatch) {
		t.Errorf("Access-Control-Allow-Methods = %q, want PATCH", got)
	}
}

func TestCorsMiddlewareSupportsMultipleAllowedOrigins(t *testing.T) {
	config.App.AllowedOrigin = "http://localhost:5173,https://my-social.vercel.app"
	handler := CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	req1.Header.Set("Origin", "https://my-social.vercel.app")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if got := rec1.Header().Get("Access-Control-Allow-Origin"); got != "https://my-social.vercel.app" {
		t.Errorf("expected Access-Control-Allow-Origin to be vercel domain, got %q", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	req2.Header.Set("Origin", "http://localhost:5173")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if got := rec2.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin to be localhost, got %q", got)
	}
}
