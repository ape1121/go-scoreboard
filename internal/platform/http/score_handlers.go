package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"

	"github.com/ape1121/go-scoreboard/internal/score"
	"github.com/go-chi/chi/v5"
)

type scoreHandler struct {
	service *score.Service
}

func newScoreHandler(service *score.Service) scoreHandler {
	return scoreHandler{service: service}
}

func (h scoreHandler) upsert(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var request setScoreRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "invalid request body")
		return
	}

	entry, err := h.service.Set(r.Context(), request.toInput(chi.URLParam(r, "boardId")))
	if err != nil {
		writeScoreError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, toSetScoreResponse(entry))
}

func (h scoreHandler) top(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	limit, err := scoreLimit(r)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "invalid value for n")
		return
	}

	entries, err := h.service.Top(r.Context(), chi.URLParam(r, "boardId"), limit)
	if err != nil {
		writeScoreError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, toTopScoresResponse(entries))
}

func (h scoreHandler) surroundings(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	_ = h.service
	writeError(w, stdhttp.StatusNotImplemented, "endpoint not implemented")
	_ = r
}

func writeScoreError(w stdhttp.ResponseWriter, err error) {
	var validationErr score.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(w, stdhttp.StatusBadRequest, validationErr.Error())
	case errors.Is(err, score.ErrBoardNotFound):
		writeError(w, stdhttp.StatusNotFound, "board not found")
	default:
		writeError(w, stdhttp.StatusInternalServerError, "internal server error")
	}
}

func scoreLimit(r *stdhttp.Request) (int, error) {
	const defaultLimit = 10

	raw := r.URL.Query().Get("n")
	if raw == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}

	return limit, nil
}
