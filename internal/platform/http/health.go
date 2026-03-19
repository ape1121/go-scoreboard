package http

import (
	"context"
	stdhttp "net/http"
)

type Pinger interface {
	Ping(context.Context) error
}

type HealthService struct {
	pinger Pinger
}

func NewHealthService(pinger Pinger) HealthService {
	return HealthService{pinger: pinger}
}

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
