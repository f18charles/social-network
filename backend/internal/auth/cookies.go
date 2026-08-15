package auth

import (
	"errors"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

const (
	CookieName = "session_id"
	Session    = 24 * time.Hour
)

func sessionCookieAttributes() (bool, http.SameSite) {
	secure := strings.EqualFold(os.Getenv("APP_ENV"), "production")
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	if val := os.Getenv("COOKIE_SECURE"); val != "" {
		secure = strings.EqualFold(val, "true") || val == "1"
	}
	if val := os.Getenv("COOKIE_SAME_SITE"); val != "" {
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "lax":
			sameSite = http.SameSiteLaxMode
		case "strict":
			sameSite = http.SameSiteStrictMode
		case "none":
			sameSite = http.SameSiteNoneMode
		}
	}
	return secure, sameSite
}

// SetCookie creates the session cookie for a logged in user
func SetCookie(w http.ResponseWriter, session_id uuid.UUID) {
	secure, sameSite := sessionCookieAttributes()
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    session_id.String(),
		Path:     "/",
		MaxAge:   math.MaxInt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

// ClearCookie destroys the session cookie when the user logs out
func ClearCookie(w http.ResponseWriter) {
	secure, sameSite := sessionCookieAttributes()
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
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
