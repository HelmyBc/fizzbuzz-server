package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// errorResponse is the JSON body returned for API errors.
type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The status and headers have already been sent, so we cannot
		// change the response. Log the encoding error instead.
		slog.Error("failed to encode response JSON", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
