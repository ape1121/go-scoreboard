package http

import (
	"context"
	stdhttp "net/http"
)

// Pinger captures the minimum database behavior needed for readiness checks.
type Pinger interface {
	Ping(context.Context) error
}

// HealthService exposes dependencies needed by readiness checks.
type HealthService struct {
	pinger Pinger
}

// NewHealthService wires the dependencies used by the health endpoint.
func NewHealthService(pinger Pinger) HealthService {
	return HealthService{pinger: pinger}
}

// Ready reports whether critical dependencies are reachable.
func (s HealthService) Ready(ctx context.Context) error {
	return s.pinger.Ping(ctx)
}

func healthHandler(service HealthService) stdhttp.HandlerFunc {
	type response struct {
		Status string `json:"status"`
	}

	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if err := service.Ready(r.Context()); err != nil {
			writeJSON(w, stdhttp.StatusServiceUnavailable, response{Status: "unhealthy"})
			return
		}

		writeJSON(w, stdhttp.StatusOK, response{Status: "ok"})
	}
}
