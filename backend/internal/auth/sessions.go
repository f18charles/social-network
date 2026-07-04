package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

// Create creates a sessiona and stores it in the database
// It also creates the session cookie
func Create(db *sql.DB, w http.ResponseWriter, user_id uuid.UUID) error {
	var new_session models.Session

	new_session = models.Session{
		ID:        uuid.Must(uuid.NewV4()),
		UserID:    user_id,
		ExpiresAt: time.Now().AddDate(100, 0, 0),
		CreatedAt: time.Now(),
	}

	repo := repositories.NewSessionRepository(db)
	if err := repo.CreateSession(&new_session); err != nil {
		return err
	}

	SetCookie(w, new_session.ID)
	return nil
}

// Destroy removes an existing cookie and deletes an ongoing session
func Destroy(db *sql.DB, w http.ResponseWriter, r *http.Request) error {
	token, err := GetCookieValue(r)
	if err != nil {
		ClearCookie(w)
		return nil
	}

	repo := repositories.NewSessionRepository(db)
	err = repo.DeleteSession(token)
	ClearCookie(w)

	return err
}

// Validate checks whether a session is active or expired
// It returns a user ID as a string, or an error if the
// session doesn't exist or has expired.
func Validate(db *sql.DB, token uuid.UUID) (string, error) {
	repo := repositories.NewSessionRepository(db)
	s, err := repo.GetSessionByID(token)
	if err != nil {
		return "", err
	}

	if time.Now().After(s.ExpiresAt) {
		if err := repo.DeleteSession(s.ID); err != nil {
			return "", err
		}
		return "", errors.New("session expired")
	}

	return s.UserID.String(), nil
}
