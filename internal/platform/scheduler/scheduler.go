package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

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

type Repository interface {
	DueBoardIDs(context.Context, time.Time) ([]string, error)
	ResetDueBoard(context.Context, string, time.Time) (bool, error)
}

type DBRunner struct {
	repository Repository
}

func NewRunner(repository Repository) DBRunner {
	return DBRunner{repository: repository}
}

func (r DBRunner) Run(ctx context.Context, now time.Time) error {
	boardIDs, err := r.repository.DueBoardIDs(ctx, now)
	if err != nil {
		return fmt.Errorf("load due boards: %w", err)
	}

	var runErr error
	for _, boardID := range boardIDs {
		if _, err := r.repository.ResetDueBoard(ctx, boardID, now); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("reset board %s: %w", boardID, err))
		}
	}

	return runErr
}

func New(logger *log.Logger, pollInterval time.Duration, serviceClock clock.Clock, runner Runner) *Scheduler {
	if serviceClock == nil {
		serviceClock = clock.System{}
	}

	return &Scheduler{
		logger:       logger,
		pollInterval: pollInterval,
		clock:        serviceClock,
		runner:       runner,
	}
}

func (s *Scheduler) CatchUp(ctx context.Context) error {
	return s.runner.Run(ctx, s.clock.Now())
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
