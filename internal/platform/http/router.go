package http

import (
	"log"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"
)

// Dependencies captures the HTTP-layer collaborators.
type Dependencies struct {
	Logger        *log.Logger
	HealthService HealthService
}

// NewRouter assembles the public HTTP routes.
func NewRouter(deps Dependencies) stdhttp.Handler {
	router := chi.NewRouter()

	_ = deps.Logger

	router.Get("/healthz", healthHandler(deps.HealthService))

	return router
}
