package middleware

import (
	"context"
	"net/http"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

type authContextKey string

const UserContextKey authContextKey = "user"

func Auth(userService services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized: Session cookie missing", nil)
				return
			}

			user, err := userService.Authenticate(cookie.Value)
			if err != nil {
				_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized: Invalid or expired session", nil)
				return
			}

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
