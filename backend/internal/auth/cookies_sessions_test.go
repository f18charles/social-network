package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	dbpkg "learn.zone01kisumu.ke/git/qquinton/social-network/internal/db"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

func TestSetGetAndClearCookie(t *testing.T) {
	sessionID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000201"))
	recorder := httptest.NewRecorder()

	SetCookie(recorder, sessionID)

	response := recorder.Result()
	defer response.Body.Close()
	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != CookieName || cookie.Value != sessionID.String() || !cookie.HttpOnly {
		t.Fatalf("cookie = %#v", cookie)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookie)
	got, err := GetCookieValue(request)
	if err != nil {
		t.Fatalf("GetCookieValue returned error: %v", err)
	}
	if got != sessionID {
		t.Fatalf("cookie value = %s, want %s", got, sessionID)
	}

	clearRecorder := httptest.NewRecorder()
	ClearCookie(clearRecorder)
	clearCookie := clearRecorder.Result().Cookies()[0]
	if clearCookie.Value != "" || clearCookie.MaxAge >= 0 {
		t.Fatalf("clear cookie = %#v", clearCookie)
	}
}

func TestGetCookieValueRejectsMissingEmptyAndMalformedCookies(t *testing.T) {
	if _, err := GetCookieValue(httptest.NewRequest(http.MethodGet, "/", nil)); err == nil {
		t.Fatal("expected missing cookie to fail")
	}

	emptyRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	emptyRequest.AddCookie(&http.Cookie{Name: CookieName, Value: ""})
	if _, err := GetCookieValue(emptyRequest); err == nil {
		t.Fatal("expected empty cookie to fail")
	}

	badRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	badRequest.AddCookie(&http.Cookie{Name: CookieName, Value: "not-a-uuid"})
	if _, err := GetCookieValue(badRequest); err == nil {
		t.Fatal("expected malformed cookie to fail")
	}
}

func TestCreateValidateAndDestroySession(t *testing.T) {
	db := newAuthTestDB(t)
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000201"))
	insertAuthTestUser(t, db, userID)

	createRecorder := httptest.NewRecorder()
	if err := Create(db, createRecorder, userID); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	cookies := createRecorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	sessionID, err := uuid.FromString(cookies[0].Value)
	if err != nil {
		t.Fatalf("created cookie value is not a uuid: %v", err)
	}

	gotUserID, err := Validate(db, sessionID)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if gotUserID != userID.String() {
		t.Fatalf("validated user id = %s, want %s", gotUserID, userID)
	}

	destroyRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
	destroyRequest.AddCookie(cookies[0])
	destroyRecorder := httptest.NewRecorder()
	if err := Destroy(db, destroyRecorder, destroyRequest); err != nil {
		t.Fatalf("Destroy returned error: %v", err)
	}
	if _, err := Validate(db, sessionID); err == nil {
		t.Fatal("expected destroyed session validation to fail")
	}
	clearCookie := destroyRecorder.Result().Cookies()[0]
	if clearCookie.Value != "" || clearCookie.MaxAge >= 0 {
		t.Fatalf("destroy clear cookie = %#v", clearCookie)
	}
}

func TestValidateDeletesExpiredSessions(t *testing.T) {
	db := newAuthTestDB(t)
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000202"))
	sessionID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000202"))
	insertAuthTestUser(t, db, userID)
	repo := repositories.NewSessionRepository(db)
	if err := repo.CreateSession(&models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(-time.Minute),
		CreatedAt: time.Now().Add(-2 * time.Minute),
	}); err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	if _, err := Validate(db, sessionID); err == nil {
		t.Fatal("expected expired session validation to fail")
	}
	if _, err := repo.GetSessionByID(sessionID); err == nil {
		t.Fatal("expected expired session to be deleted")
	}
}

func newAuthTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := dbpkg.InitDB(filepath.Join(t.TempDir(), "auth.db"), filepath.Join("..", "db", "migrations"), "", "")
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func insertAuthTestUser(t *testing.T, db *sql.DB, id uuid.UUID) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, dob,
			avatar, nickname, about_me, is_public, created_at
		)
		VALUES (?, ?, 'hash', 'Auth', 'User', '1998-04-12', NULL, NULL, NULL, 1, ?)
	`, id.String(), id.String()+"@example.com", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}
