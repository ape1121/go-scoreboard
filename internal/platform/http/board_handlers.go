package http

import (
	stdhttp "net/http"

	"github.com/ape1121/go-scoreboard/internal/board"
)

type boardHandler struct {
	service *board.Service
}

func newBoardHandler(service *board.Service) boardHandler {
	return boardHandler{service: service}
}

func (h boardHandler) create(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	_ = h.service
	writeError(w, stdhttp.StatusNotImplemented, "endpoint not implemented")
	_ = r
}

func (h boardHandler) list(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	_ = h.service
	writeError(w, stdhttp.StatusNotImplemented, "endpoint not implemented")
	_ = r
}

func (h boardHandler) get(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	_ = h.service
	writeError(w, stdhttp.StatusNotImplemented, "endpoint not implemented")
	_ = r
}
