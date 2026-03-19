package http

import (
	"log"
	stdhttp "net/http"
	"time"

	"github.com/ape1121/go-scoreboard/internal/platform/config"
)

// NewServer creates the HTTP server with conservative defaults suitable for APIs.
func NewServer(cfg config.Config, logger *log.Logger, handler stdhttp.Handler) *stdhttp.Server {
	return &stdhttp.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           handler,
		ErrorLog:          logger,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
