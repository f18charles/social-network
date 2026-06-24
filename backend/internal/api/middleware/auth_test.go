package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

func TestAuthRejectsMissingSessionCookie(t *testing.T) {
	service := &fakeAuthUserService{}
	handler := Auth(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAuthRejectsInvalidSession(t *testing.T) {
	service := &fakeAuthUserService{authErr: errAuthTest}
	handler := Auth(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	request.AddCookie(&http.Cookie{Name: "session_token", Value: "bad-token"})

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if service.lastSessionID != "bad-token" {
		t.Fatalf("Authenticate called with %q, want bad-token", service.lastSessionID)
	}
}

func TestAuthAddsUserToContext(t *testing.T) {
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000301"))
	service := &fakeAuthUserService{user: &models.User{ID: userID, Email: "viewer@example.com"}}
	handler := Auth(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			t.Fatal("expected user in context")
		}
		if user.ID != userID {
			t.Fatalf("context user = %s, want %s", user.ID, userID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	request.AddCookie(&http.Cookie{Name: "session_token", Value: "good-token"})

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if service.lastSessionID != "good-token" {
		t.Fatalf("Authenticate called with %q, want good-token", service.lastSessionID)
	}
}

type authTestError string

func (e authTestError) Error() string { return string(e) }

const errAuthTest = authTestError("auth failed")

type fakeAuthUserService struct {
	user          *models.User
	authErr       error
	lastSessionID string
}

func (s *fakeAuthUserService) Register(req *models.CreateUserRequest) (*models.UserResponse, error) {
	return nil, nil
}

func (s *fakeAuthUserService) Login(email, password string) (*models.Session, error) {
	return nil, nil
}

func (s *fakeAuthUserService) Logout(sessionID string) error {
	return nil
}

func (s *fakeAuthUserService) Authenticate(sessionID string) (*models.User, error) {
	s.lastSessionID = sessionID
	if s.authErr != nil {
		return nil, s.authErr
	}
	return s.user, nil
}

func (s *fakeAuthUserService) GetByID(id uuid.UUID) (*models.User, error) {
	return nil, nil
}

func (s *fakeAuthUserService) ListAllUsers(query string, excludeID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}

func (s *fakeAuthUserService) Update(userID uuid.UUID, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	return &models.UserResponse{ID: userID, CreatedAt: time.Now()}, nil
}
