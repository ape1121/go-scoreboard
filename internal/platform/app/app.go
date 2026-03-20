package app

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ape1121/go-scoreboard/internal/board"
	boardpostgres "github.com/ape1121/go-scoreboard/internal/board/postgres"
	"github.com/ape1121/go-scoreboard/internal/platform/clock"
	"github.com/ape1121/go-scoreboard/internal/platform/config"
	platformdb "github.com/ape1121/go-scoreboard/internal/platform/db"
	platformhttp "github.com/ape1121/go-scoreboard/internal/platform/http"
	"github.com/ape1121/go-scoreboard/internal/platform/scheduler"
	"github.com/ape1121/go-scoreboard/internal/score"
	scorepostgres "github.com/ape1121/go-scoreboard/internal/score/postgres"
)

type Application struct {
	Server    *http.Server
	Scheduler *scheduler.Scheduler
}

func Build(ctx context.Context, cfg config.Config, logger *log.Logger) (*Application, func(), error) {
	pool, err := platformdb.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		pool.Close()
	}

	systemClock := clock.System{}
	boardRepository := boardpostgres.NewRepository(pool)
	scoreRepository := scorepostgres.NewRepository(pool)
	boardService := board.NewService(boardRepository, systemClock, board.NewID)
	scoreService := score.NewService(scoreRepository, boardRepository, systemClock)
	healthService := platformhttp.NewHealthService(pool)
	httpDeps := platformhttp.Dependencies{
		Logger:        logger,
		HealthService: healthService,
		BoardService:  boardService,
		ScoreService:  scoreService,
	}

	return &Application{
		Server:    platformhttp.NewServer(cfg, logger, platformhttp.NewRouter(httpDeps)),
		Scheduler: scheduler.New(logger, cfg.SchedulerPollInterval, systemClock, scheduler.NewRunner(boardRepository)),
	}, cleanup, nil
}

var _ board.Repository = (*boardpostgres.Repository)(nil)
var _ scheduler.Repository = (*boardpostgres.Repository)(nil)
var _ score.BoardResolver = (*boardpostgres.Repository)(nil)
var _ score.Repository = (*scorepostgres.Repository)(nil)
var _ platformhttp.Pinger = (*pgxpool.Pool)(nil)
