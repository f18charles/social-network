package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

type loggingContextKey string

const RequestIDKey loggingContextKey = "request_id"

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(RequestIDKey).(string); ok {
		return val
	}
	return ""
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code and bytes written.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.written += int64(n)
	return n, err
}

// RequestTracingAndLoggingMiddleware traces and logs incoming HTTP requests.
func RequestTracingAndLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get or generate request ID
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			u, err := uuid.NewV4()
			if err == nil {
				reqID = u.String()
			} else {
				reqID = time.Now().Format("20060102150405.000000")
			}
		}

		// Inject into context
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		r = r.WithContext(ctx)

		// Set header in response
		w.Header().Set("X-Request-ID", reqID)

		// Wrap response writer
		wrapped := &responseWriterWrapper{ResponseWriter: w}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Determine status code (default to 200 OK if WriteHeader wasn't called)
		status := wrapped.statusCode
		if status == 0 {
			status = http.StatusOK
		}

		// Write structured log
		log.Printf("[Request] ID=%s Method=%s Path=%s Remote=%s Status=%d Size=%d Duration=%v",
			reqID, r.Method, r.URL.Path, r.RemoteAddr, status, wrapped.written, time.Since(start))
	})
}
