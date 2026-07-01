package services

import (
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

func TestUserServiceRegisterRequiresProjectRegistrationFields(t *testing.T) {
	service := NewUserService(newFakeUserRepository(), newFakeSessionRepository())

	_, err := service.Register(&dto.CreateUserRequest{
		Email:       "amina@example.com",
		Password:    "secret1",
		FirstName:   "Amina",
		DateOfBirth: "1998-04-12",
		IsPublic:    true,
	})

	if err == nil {
		t.Fatal("expected missing last name to be rejected")
	}
}

func TestUserServiceRegisterStoresHashAndReturnsSafeProfile(t *testing.T) {
	users := newFakeUserRepository()
	service := NewUserService(users, newFakeSessionRepository())

	response, err := service.Register(&dto.CreateUserRequest{
		Email:       "amina@example.com",
		Password:    "secret1",
		FirstName:   "Amina",
		LastName:    "Njeri",
		DateOfBirth: "1998-04-12",
		Avatar:      "/uploads/avatars/amina.png",
		Nickname:    "amina",
		AboutMe:     "Weekend hiker",
		IsPublic:    true,
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	stored := users.byEmail["amina@example.com"]
	if stored == nil {
		t.Fatal("expected user to be stored")
	}
	if stored.PassHash == "secret1" {
		t.Fatal("expected password to be hashed, not stored as plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PassHash), []byte("secret1")); err != nil {
		t.Fatalf("stored hash does not match password: %v", err)
	}
	if response.Email != "amina@example.com" || response.FirstName != "Amina" || response.LastName != "Njeri" {
		t.Fatalf("unexpected response profile: %#v", response)
	}
}

func TestUserServiceRegisterRejectsDuplicateEmail(t *testing.T) {
	users := newFakeUserRepository()
	service := NewUserService(users, newFakeSessionRepository())
	users.add(&models.User{
		ID:        uuid.Must(uuid.NewV4()),
		Email:     "amina@example.com",
		PassHash:  "hash",
		FirstName: "Amina",
		LastName:  "Njeri",
	})

	_, err := service.Register(&dto.CreateUserRequest{
		Email:       "amina@example.com",
		Password:    "secret1",
		FirstName:   "Amina",
		LastName:    "Njeri",
		DateOfBirth: "1998-04-12",
		IsPublic:    true,
	})

	if err == nil || err.Error() != "email already registered" {
		t.Fatalf("error = %v, want duplicate email error", err)
	}
}

func TestUserServiceLoginCreatesSessionForValidPassword(t *testing.T) {
	users := newFakeUserRepository()
	sessions := newFakeSessionRepository()
	service := NewUserService(users, sessions)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword returned error: %v", err)
	}
	userID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
	users.add(&models.User{
		ID:        userID,
		Email:     "amina@example.com",
		PassHash:  string(hash),
		FirstName: "Amina",
		LastName:  "Njeri",
		IsPublic:  true,
	})

	session, err := service.Login("amina@example.com", "secret1")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if session.UserID != userID {
		t.Fatalf("session user id = %s, want %s", session.UserID, userID)
	}
	if sessions.byID[session.ID] == nil {
		t.Fatal("expected created session to be persisted")
	}
}

func TestUserServiceAuthenticateRejectsExpiredSessionAndDeletesIt(t *testing.T) {
	users := newFakeUserRepository()
	sessions := newFakeSessionRepository()
	service := NewUserService(users, sessions)
	userID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
	sessionID := uuid.Must(uuid.FromString("0dd6e443-0998-4f50-a4cf-1a40a0536213"))
	users.add(&models.User{ID: userID, Email: "amina@example.com"})
	sessions.byID[sessionID] = &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(-time.Minute),
	}

	_, err := service.Authenticate(sessionID.String())

	if err == nil || err.Error() != "session expired" {
		t.Fatalf("error = %v, want session expired", err)
	}
	if sessions.byID[sessionID] != nil {
		t.Fatal("expected expired session to be deleted")
	}
}

func TestUserServiceLoginRejectsInvalidPassword(t *testing.T) {
	users := newFakeUserRepository()
	service := NewUserService(users, newFakeSessionRepository())

	hash, err := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword returned error: %v", err)
	}
	users.add(&models.User{
		ID:        uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1")),
		Email:     "amina@example.com",
		PassHash:  string(hash),
		FirstName: "Amina",
		LastName:  "Njeri",
		IsPublic:  true,
	})

	if _, err := service.Login("amina@example.com", "wrong-password"); err == nil {
		t.Fatal("expected invalid password to be rejected")
	}
}

func TestUserServiceUpdateRequiresCurrentPasswordForSensitiveChanges(t *testing.T) {
	users := newFakeUserRepository()
	service := NewUserService(users, newFakeSessionRepository())
	hash, err := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword returned error: %v", err)
	}
	userID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
	users.add(&models.User{
		ID:        userID,
		Email:     "amina@example.com",
		PassHash:  string(hash),
		FirstName: "Amina",
		LastName:  "Njeri",
		DOB:       time.Date(1998, 4, 12, 0, 0, 0, 0, time.UTC),
		IsPublic:  true,
	})

	if _, err := service.Update(userID, &dto.UpdateUserRequest{Email: "new@example.com"}); err == nil {
		t.Fatal("expected email update without current password to fail")
	}
	if _, err := service.Update(userID, &dto.UpdateUserRequest{NewPassword: "secret2"}); err == nil {
		t.Fatal("expected password update without current password to fail")
	}
}

func TestUserServiceUpdateChangesProfileAndPassword(t *testing.T) {
	users := newFakeUserRepository()
	service := NewUserService(users, newFakeSessionRepository())
	hash, err := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword returned error: %v", err)
	}
	userID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
	users.add(&models.User{
		ID:        userID,
		Email:     "amina@example.com",
		PassHash:  string(hash),
		FirstName: "Amina",
		LastName:  "Njeri",
		DOB:       time.Date(1998, 4, 12, 0, 0, 0, 0, time.UTC),
		IsPublic:  true,
		CreatedAt: time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
	})

	response, err := service.Update(userID, &dto.UpdateUserRequest{
		Email:           "new@example.com",
		CurrentPassword: "secret1",
		NewPassword:     "secret2",
		FirstName:       "New",
		LastName:        "Name",
		DateOfBirth:     "1999-05-13",
		Nickname:        "nn",
		AboutMe:         "updated",
		Avatar:          "/uploads/avatars/new.png",
		IsPublic:        false,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	stored := users.byID[userID]
	if response.Email != "new@example.com" || response.FirstName != "New" || response.DateOfBirth != "1999-05-13" {
		t.Fatalf("unexpected response: %#v", response)
	}
	if stored.Nickname != "nn" || stored.AboutMe != "updated" || stored.Avatar != "/uploads/avatars/new.png" || stored.IsPublic {
		t.Fatalf("profile fields not updated: %#v", stored)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PassHash), []byte("secret2")); err != nil {
		t.Fatalf("updated password hash does not match: %v", err)
	}
}

type fakeUserRepository struct {
	byID    map[uuid.UUID]*models.User
	byEmail map[string]*models.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		byID:    map[uuid.UUID]*models.User{},
		byEmail: map[string]*models.User{},
	}
}

func (r *fakeUserRepository) add(user *models.User) {
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
}

func (r *fakeUserRepository) CreateUser(user *models.User) error {
	r.add(user)
	return nil
}

func (r *fakeUserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	user := r.byID[id]
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *fakeUserRepository) GetUserByEmail(email string) (*models.User, error) {
	user := r.byEmail[email]
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *fakeUserRepository) ListUsers(query string, excludeID uuid.UUID) ([]*models.User, error) {
	return []*models.User{}, nil
}

func (r *fakeUserRepository) UpdateUserProfile(user *models.User) error {
	r.add(user)
	return nil
}

func (r *fakeUserRepository) DeleteUser(id uuid.UUID) error {
	delete(r.byID, id)
	return nil
}

type fakeSessionRepository struct {
	byID map[uuid.UUID]*models.Session
}

func newFakeSessionRepository() *fakeSessionRepository {
	return &fakeSessionRepository{byID: map[uuid.UUID]*models.Session{}}
}

func (r *fakeSessionRepository) CreateSession(session *models.Session) error {
	r.byID[session.ID] = session
	return nil
}

func (r *fakeSessionRepository) GetSessionByID(id uuid.UUID) (*models.Session, error) {
	session := r.byID[id]
	if session == nil {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (r *fakeSessionRepository) DeleteSession(id uuid.UUID) error {
	delete(r.byID, id)
	return nil
}
