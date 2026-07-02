package auth

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
)

const (
	CookieName = "session_id"
	Session    = 24 * time.Hour
)

// SetCookie creates the session cookie for a logged in user
func SetCookie(w http.ResponseWriter, session_id uuid.UUID) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    session_id.String(),
		Path:     "/",
		MaxAge:   math.MaxInt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearCookie destroys the session cookie when the user logs out
func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetCookieValue confirms if a session cookie exists
func GetCookieValue(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return uuid.Nil, errors.New("no session cookie")
	}
	if cookie.Value == "" {
		return uuid.Nil, errors.New("empty session cookie")
	}
	return uuid.FromString(cookie.Value)
}
