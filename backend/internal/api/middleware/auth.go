package middleware

import (
	"context"
	"net/http"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

// Auth checks the request's session cookie, validates the session id and attaches
// the authenticated user to the request so that the program knows the current auth
// state of the user
func Auth(userService services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Pull the session cookie from the incoming request
			cookie, err := r.Cookie("session_token")
			if err != nil {
				_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized: Session cookie missing", nil)
				return
			}

			// validate the session id(cookie.Value) against the session store
			user, err := userService.Authenticate(cookie.Value)
			if err != nil {
				_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized: Invalid or expired session", nil)
				return
			}

			// attaches the authenticated user to the request
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext helper to retrieve user model from context.
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(UserContextKey).(*models.User)
	return u, ok
}
