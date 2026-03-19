package http

import (
	stdhttp "net/http"

	"github.com/ape1121/go-scoreboard/internal/score"
)

type scoreHandler struct {
	service *score.Service
}

func newScoreHandler(service *score.Service) scoreHandler {
	return scoreHandler{service: service}
}

func (h scoreHandler) upsert(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	_ = h.service
	writeError(w, stdhttp.StatusNotImplemented, "endpoint not implemented")
	_ = r
}

func (h scoreHandler) top(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	_ = h.service
	writeError(w, stdhttp.StatusNotImplemented, "endpoint not implemented")
	_ = r
}

func (h scoreHandler) surroundings(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	_ = h.service
	writeError(w, stdhttp.StatusNotImplemented, "endpoint not implemented")
	_ = r
}
