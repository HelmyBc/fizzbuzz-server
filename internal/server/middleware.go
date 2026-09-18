package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type ctxKey struct{}

var requestIDCtxKey = ctxKey{}

// withRequestID reads the X-Request-ID header from the incoming request. If
// absent, it generates a cryptographically random 16-hex-character ID. The ID
// is stored in the request context (accessible via requestIDFromCtx) and
// echoed back in the X-Request-ID response header
func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = newRequestID()
			}
			w.Header().Set("X-Request-Id", id)
			ctx := context.WithValue(r.Context(), requestIDCtxKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

// requestIDFromCtx extracts the request ID from the context (stored by withRequestID)
// Returns an empty string if no ID is present
func requestIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(requestIDCtxKey).(string)
	return id
}

// newRequestID generates a random 16-hex-character request ID.
// used so IDs are unique even under high concurrency.
func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// statusRecorder is a wrapper around http.ResponseWriter that records the status code.
type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// withLogging logs one structured line per request: method, path, status
// code, duration, and request ID. This is the minimum needed to debug
// production incidents.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{
				ResponseWriter: w,
				Status:         http.StatusOK,
			}
			next.ServeHTTP(rec, r)
			duration := time.Since(start).Milliseconds()
			slog.Info(
				"request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.Status,
				"duration_ms", duration,
				"request_id", requestIDFromCtx(r.Context()),
			)
		},
	)
}

// withRecovery converts a panic in any handler into a 500 response instead
// of crashing the whole server process.
func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					slog.Error("panic", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		},
	)
}
