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

func Create(db *sql.DB, w http.ResponseWriter, user_id uuid.UUID) error {
	expires_at := time.Now().Add(Session)

	var new_session models.Session

	new_session = models.Session{
		ID:        uuid.Must(uuid.NewV4()),
		UserID:    user_id,
		ExpiresAt: expires_at,
		CreatedAt: time.Now(),
	}

	repo := repositories.NewSessionRepository(db)
	if err := repo.CreateSession(&new_session); err != nil {
		return err
	}

	SetCookie(w, new_session.ID)
	return nil
}

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
