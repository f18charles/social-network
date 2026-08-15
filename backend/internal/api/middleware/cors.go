package middleware

import (
	"log"
	"net/http"
	"strings"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/config"
)

// CorsMiddleware sets CORS headers, answers OPTIONS preflights, and logs
// each request
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := config.App.AllowedOrigin

		if origin != "" {
			if config.App.IsOriginAllowed(origin) {
				allowedOrigin = origin
			}
		} else if strings.Contains(allowedOrigin, ",") {
			// Default to first configured origin if multiple are defined
			parts := strings.Split(allowedOrigin, ",")
			if len(parts) > 0 {
				allowedOrigin = strings.TrimSpace(parts[0])
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie, X-Requested-With, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
