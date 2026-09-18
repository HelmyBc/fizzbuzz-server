package server

import (
	"net/http"

	"github.com/HelmyBc/fizzbuzz-server/internal/stats"
)

// NewRouter builds the complete HTTP handler for the service.
//
// Middleware chain (outermost → innermost):
// withRequestID → withLogging → withRecovery → mux
//
// withRequestID is outermost so every downstream middleware and handler
// can read the request ID from the context. withLogging wraps
// withRecovery so it can record the final status code after recovery
// handles a panic.
func NewRouter(store *stats.Store) http.Handler {
	mux := http.NewServeMux() // traffic director

	// Method specific routes, reached only when method matches.
	mux.HandleFunc("GET /healthz", HealthHandler)
	mux.Handle("GET /fizzbuzz", &FizzBuzzHandler{Stats: store})
	mux.Handle("GET /statistics", &StatisticsHandler{Stats: store})

	// Bare-path fallbacks, reached only when the method-specific route above
	// did NOT match, these handlers are only invoked for non-GET methods.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed, use GET")
	})
	mux.HandleFunc("/fizzbuzz", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed, use GET")
	})
	mux.HandleFunc("/statistics", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed, use GET")
	})

	// Catch-all: any path that didn't match a more specific pattern above.
	// Returns a JSON 404.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "no such endpoint")
	})

	return withRequestID(withLogging(withRecovery(mux)))
}
