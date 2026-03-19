package http

import (
	"errors"
	"io"
	stdhttp "net/http"

	"github.com/ape1121/go-scoreboard/internal/board"
	"github.com/go-chi/chi/v5"
)

type boardHandler struct {
	service *board.Service
}

func newBoardHandler(service *board.Service) boardHandler {
	return boardHandler{service: service}
}

func (h boardHandler) create(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var request createBoardRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "invalid request body")
		return
	}

	entity, err := h.service.Create(r.Context(), request.toInput())
	if err != nil {
		writeBoardError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusCreated, toCreateBoardResponse(entity))
}

func (h boardHandler) list(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	boards, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, stdhttp.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, stdhttp.StatusOK, toBoardListResponse(boards))
}

func (h boardHandler) get(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	details, err := h.service.Get(r.Context(), chi.URLParam(r, "boardId"))
	if err != nil {
		writeBoardError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, toGetBoardResponse(details))
}

func writeBoardError(w stdhttp.ResponseWriter, err error) {
	var validationErr board.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(w, stdhttp.StatusBadRequest, validationErr.Error())
	case errors.Is(err, board.ErrNotFound):
		writeError(w, stdhttp.StatusNotFound, "board not found")
	case errors.Is(err, io.EOF):
		writeError(w, stdhttp.StatusBadRequest, "invalid request body")
	default:
		writeError(w, stdhttp.StatusInternalServerError, "internal server error")
	}
}
