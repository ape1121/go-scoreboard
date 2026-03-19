package scheduler

import (
	"context"
	"log"
	"time"
)

// Scheduler is the background job coordinator for periodic board resets.
type Scheduler struct {
	logger       *log.Logger
	pollInterval time.Duration
}

// New creates a scheduler with the configured polling cadence.
func New(logger *log.Logger, pollInterval time.Duration) *Scheduler {
	return &Scheduler{
		logger:       logger,
		pollInterval: pollInterval,
	}
}

// Start runs the scheduler loop until the context is cancelled.
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
				// Phase 5 will add due-board lookup and reset execution here.
			}
		}
	}()
}
