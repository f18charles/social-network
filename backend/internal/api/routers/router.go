package routers

import (
	"log"
	"net/http"
	"os"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/handlers"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
)

// RegisterRoutes configures the application's HTTP ServeMux.
// It maps URL paths to their corresponding handler functions, sets up static file serving for uploads,
// and wraps the router with necessary middleware like CORS and authentication.
// Note that all registered application endpoints fall under the "/api/" path.
func RegisterRoutes(
	userHandler *handlers.UserHandler,
	followerHandler *handlers.FollowerHandler,
	authMiddleware func(http.Handler) http.Handler,
	allowedOrigin string,
) http.Handler {
	// Initialize ServeMux
	mux := http.NewServeMux()

	// Serve static uploads (for avatar files)
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Public routes
	mux.HandleFunc("/api/users/register", userHandler.Register)
	mux.HandleFunc("/api/users/login", userHandler.Login)

	// Authenticated routes
	mux.Handle("/api/users/me", http.HandlerFunc(userHandler.Me))
	mux.HandleFunc("/api/users/logout", userHandler.Logout)

	mux.Handle("/api/followers/follow", authMiddleware(http.HandlerFunc(followerHandler.Follow)))
	mux.Handle("/api/followers/unfollow", authMiddleware(http.HandlerFunc(followerHandler.Unfollow)))
	mux.Handle("/api/followers/accept", authMiddleware(http.HandlerFunc(followerHandler.AcceptFollow)))
	mux.Handle("/api/followers/reject", authMiddleware(http.HandlerFunc(followerHandler.RejectFollow)))
	mux.Handle("/api/followers/followers", authMiddleware(http.HandlerFunc(followerHandler.GetFollowers)))
	mux.Handle("/api/followers/following", authMiddleware(http.HandlerFunc(followerHandler.GetFollowing)))

	return middleware.CorsMiddleware(mux)
}
