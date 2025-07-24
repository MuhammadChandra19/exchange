package healthcheck

import (
	"fmt"
	"net/http"
)

// HealthCheck is the health check handler.
type HealthCheck struct {
}

// Handler is used to control the flow of GET /health endpoint
func (hc HealthCheck) Handler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if IsHealthCheckRequest(r) {
			hc.ServeHTTP(w, r)

			return
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// ServeHTTP serve http request for health check
func (hc HealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

// IsHealthCheckRequest is used to check if the request is a health check request
func IsHealthCheckRequest(r *http.Request) bool {
	return r.Method == "GET" && r.URL.Path == "/health"
}
