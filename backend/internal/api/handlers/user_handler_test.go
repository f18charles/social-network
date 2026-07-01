package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

func TestUserHandlerRegisterJSONReturnsCreatedEnvelope(t *testing.T) {
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000401"))
	service := &handlerFakeUserService{
		registerResponse: &dto.UserResponse{
			ID:          userID,
			Email:       "amina@example.com",
			FirstName:   "Amina",
			LastName:    "Njeri",
			DateOfBirth: "1998-04-12",
			IsPublic:    true,
		},
	}
	handler := newTestUserHandler(service)
	body := bytes.NewBufferString(`{
		"email":"amina@example.com",
		"password":"secret1",
		"first_name":"Amina",
		"last_name":"Njeri",
		"date_of_birth":"1998-04-12",
		"is_public":true
	}`)
	request := httptest.NewRequest(http.MethodPost, "/api/users/register", body)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.Register(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if service.lastRegisterRequest == nil || service.lastRegisterRequest.Email != "amina@example.com" {
		t.Fatalf("register request = %#v", service.lastRegisterRequest)
	}
	var response responseEnvelope
	decodeHandlerResponse(t, recorder, &response)
	if response.Status != "success" || response.Message != "User registered successfully" {
		t.Fatalf("response = %#v", response)
	}
}

func TestUserHandlerRegisterRejectsInvalidJSON(t *testing.T) {
	handler := newTestUserHandler(&handlerFakeUserService{})
	request := httptest.NewRequest(http.MethodPost, "/api/users/register", bytes.NewBufferString(`{`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.Register(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestUserHandlerLoginSetsSessionCookie(t *testing.T) {
	sessionID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000401"))
	service := &handlerFakeUserService{
		loginSession: &models.Session{
			ID:        sessionID,
			UserID:    uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000401")),
			ExpiresAt: time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC),
		},
	}
	handler := newTestUserHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBufferString(`{"email":"amina@example.com","password":"secret1"}`))
	recorder := httptest.NewRecorder()

	handler.Login(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.lastLoginEmail != "amina@example.com" || service.lastLoginPassword != "secret1" {
		t.Fatalf("login args = %q/%q", service.lastLoginEmail, service.lastLoginPassword)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session_token" || cookies[0].Value != sessionID.String() || !cookies[0].HttpOnly {
		t.Fatalf("cookies = %#v", cookies)
	}
}

func TestUserHandlerLogoutClearsCookieAndCallsService(t *testing.T) {
	service := &handlerFakeUserService{}
	handler := newTestUserHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/api/users/logout", nil)
	request.AddCookie(&http.Cookie{Name: "session_token", Value: "token-to-delete"})
	recorder := httptest.NewRecorder()

	handler.Logout(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if service.lastLogoutSessionID != "token-to-delete" {
		t.Fatalf("logout session = %q, want token-to-delete", service.lastLogoutSessionID)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session_token" || cookies[0].Value != "" || !cookies[0].Expires.Before(time.Now()) {
		t.Fatalf("clear cookie = %#v", cookies)
	}
}

func TestUserHandlerUpdateRequiresSessionCookie(t *testing.T) {
	handler := newTestUserHandler(&handlerFakeUserService{})
	request := httptest.NewRequest(http.MethodPatch, "/api/users/update", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()

	handler.Update(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestUserHandlerUpdateAuthenticatesAndPassesPatch(t *testing.T) {
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000402"))
	service := &handlerFakeUserService{
		authUser:       &models.User{ID: userID, Email: "viewer@example.com"},
		updateResponse: &dto.UserResponse{ID: userID, Email: "new@example.com", FirstName: "New"},
	}
	handler := newTestUserHandler(service)
	request := httptest.NewRequest(http.MethodPatch, "/api/users/update", bytes.NewBufferString(`{"email":"new@example.com","first_name":"New"}`))
	request.AddCookie(&http.Cookie{Name: "session_token", Value: "session-id"})
	recorder := httptest.NewRecorder()

	handler.Update(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.lastAuthSessionID != "session-id" {
		t.Fatalf("auth session = %q, want session-id", service.lastAuthSessionID)
	}
	if service.lastUpdateUserID != userID || service.lastUpdateRequest == nil || service.lastUpdateRequest.Email != "new@example.com" {
		t.Fatalf("update args = %s %#v", service.lastUpdateUserID, service.lastUpdateRequest)
	}
}

type responseEnvelope struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    json.RawMessage   `json:"data"`
	Errors  map[string]string `json:"errors"`
}

func decodeHandlerResponse(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(recorder.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
}

func newTestUserHandler(service *handlerFakeUserService) *UserHandler {
	return NewUserHandler(service, &handlerFakeFollowerService{})
}

type handlerFakeUserService struct {
	registerResponse    *dto.UserResponse
	registerErr         error
	loginSession        *models.Session
	loginErr            error
	authUser            *models.User
	authErr             error
	getByIDUser         *models.User
	getByIDErr          error
	updateResponse      *dto.UserResponse
	updateErr           error
	lastRegisterRequest *dto.CreateUserRequest
	lastLoginEmail      string
	lastLoginPassword   string
	lastLogoutSessionID string
	lastAuthSessionID   string
	lastGetByID         uuid.UUID
	lastUpdateUserID    uuid.UUID
	lastUpdateRequest   *dto.UpdateUserRequest
}

func (s *handlerFakeUserService) Register(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	s.lastRegisterRequest = req
	if s.registerErr != nil {
		return nil, s.registerErr
	}
	return s.registerResponse, nil
}

func (s *handlerFakeUserService) Login(email, password string) (*models.Session, error) {
	s.lastLoginEmail = email
	s.lastLoginPassword = password
	if s.loginErr != nil {
		return nil, s.loginErr
	}
	return s.loginSession, nil
}

func (s *handlerFakeUserService) Logout(sessionID string) error {
	s.lastLogoutSessionID = sessionID
	return nil
}

func (s *handlerFakeUserService) Authenticate(sessionID string) (*models.User, error) {
	s.lastAuthSessionID = sessionID
	if s.authErr != nil {
		return nil, s.authErr
	}
	if s.authUser == nil {
		return nil, errors.New("unauthorized")
	}
	return s.authUser, nil
}

func (s *handlerFakeUserService) GetByID(id uuid.UUID) (*models.User, error) {
	s.lastGetByID = id
	if s.getByIDErr != nil {
		return nil, s.getByIDErr
	}
	if s.getByIDUser == nil {
		return nil, errors.New("user not found")
	}
	return s.getByIDUser, nil
}

func (s *handlerFakeUserService) ListAllUsers(query string, excludeID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}

func (s *handlerFakeUserService) Update(userID uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	s.lastUpdateUserID = userID
	s.lastUpdateRequest = req
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	return s.updateResponse, nil
}
