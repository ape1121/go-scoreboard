package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ape1121/go-scoreboard/internal/platform/clock"
)

type Scheduler struct {
	logger       *log.Logger
	pollInterval time.Duration
	clock        clock.Clock
	runner       Runner
}

type Runner interface {
	Run(context.Context, time.Time) error
}

type NoopRunner struct {
	pool *pgxpool.Pool
}

func NewNoopRunner(pool *pgxpool.Pool) NoopRunner {
	return NoopRunner{pool: pool}
}

func (r NoopRunner) Run(context.Context, time.Time) error {
	_ = r.pool
	return nil
}

func New(logger *log.Logger, pollInterval time.Duration, clock clock.Clock, runner Runner) *Scheduler {
	return &Scheduler{
		logger:       logger,
		pollInterval: pollInterval,
		clock:        clock,
		runner:       runner,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()

		s.logger.Printf("scheduler started with poll interval %s", s.pollInterval)

		for {
			select {
			case <-ctx.Done():
				s.logger.Print("scheduler stopped")
				return
			case <-ticker.C:
				if err := s.runner.Run(ctx, s.clock.Now()); err != nil {
					s.logger.Printf("scheduler tick failed: %v", err)
				}
			}
		}
	}()
}
