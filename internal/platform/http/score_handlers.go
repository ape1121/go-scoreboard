package http

import (
	"errors"
	"fmt"
	"math/rand/v2"
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
	limit, err := scoreLimit(r)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, "invalid value for n")
		return
	}

	entries, err := h.service.Surroundings(
		r.Context(),
		chi.URLParam(r, "boardId"),
		chi.URLParam(r, "userId"),
		limit,
	)
	if err != nil {
		writeScoreError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, toSurroundingsResponse(entries))
}

func (h scoreHandler) seed(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var request seedRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		request = seedRequest{Count: 20, MaxScore: 10000}
	}
	if request.Count <= 0 || request.Count > 1000 {
		request.Count = 20
	}
	if request.MaxScore <= 0 {
		request.MaxScore = 10000
	}

	boardID := chi.URLParam(r, "boardId")
	created := 0
	for i := range request.Count {
		userID := fmt.Sprintf("player_%d", i+1)
		_, err := h.service.Set(r.Context(), score.SetInput{
			BoardID: boardID,
			UserID:  userID,
			Score:   rand.Int64N(request.MaxScore) + 1,
		})
		if err != nil {
			writeScoreError(w, err)
			return
		}
		created++
	}

	writeJSON(w, stdhttp.StatusCreated, seedResponse{Created: created})
}

func writeScoreError(w stdhttp.ResponseWriter, err error) {
	var validationErr score.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(w, stdhttp.StatusBadRequest, validationErr.Error())
	case errors.Is(err, score.ErrBoardNotFound):
		writeError(w, stdhttp.StatusNotFound, "board not found")
	case errors.Is(err, score.ErrScoreNotFound):
		writeError(w, stdhttp.StatusNotFound, "score not found for user")
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
