package http

import (
	"io"
	"log"
	stdhttp "net/http"

	"github.com/ape1121/go-scoreboard/internal/board"
	"github.com/ape1121/go-scoreboard/internal/score"
	"github.com/go-chi/chi/v5"
)

type Dependencies struct {
	Logger        *log.Logger
	HealthService HealthService
	BoardService  *board.Service
	ScoreService  *score.Service
}

func NewRouter(deps Dependencies) stdhttp.Handler {
	router := chi.NewRouter()
	logger := deps.Logger
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	boardHandler := newBoardHandler(deps.BoardService)
	scoreHandler := newScoreHandler(deps.ScoreService)

	router.Use(requestIDMiddleware)
	router.Use(requestLoggerMiddleware(logger))
	router.Use(recovererMiddleware(logger))
	router.NotFound(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		writeError(w, stdhttp.StatusNotFound, "route not found")
	})
	router.MethodNotAllowed(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		writeError(w, stdhttp.StatusMethodNotAllowed, "method not allowed")
	})

	router.Get("/healthz", healthHandler(deps.HealthService))
	router.Route("/boards", func(r chi.Router) {
		r.Post("/", boardHandler.create)
		r.Get("/", boardHandler.list)
		r.Route("/{boardId}", func(r chi.Router) {
			r.Get("/", boardHandler.get)
			r.Route("/scores", func(r chi.Router) {
				r.Post("/", scoreHandler.upsert)
				r.Get("/", scoreHandler.top)
				r.Post("/seed", scoreHandler.seed)
				r.Get("/{userId}/surroundings", scoreHandler.surroundings)
			})
		})
	})

	return router
}
