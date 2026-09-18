package server

import (
	"net/http"
)

// HealthHandler serves GET /healthz. Used by orchestrators (Docker,
// Kubernetes, load balancers) to check the process is up and accepting
// connections.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
