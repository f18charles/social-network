package handlers

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

func SafeUploadsHandler(root string) http.Handler {
	fileServer := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decodedPath, err := url.PathUnescape(r.URL.Path)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		cleanPath := path.Clean("/" + decodedPath)
		if strings.Contains(decodedPath, "..") || strings.Contains(cleanPath, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}
