package scheduler

import (
	"context"
	"log"
	"time"
)

type Scheduler struct {
	logger       *log.Logger
	pollInterval time.Duration
}

func New(logger *log.Logger, pollInterval time.Duration) *Scheduler {
	return &Scheduler{
		logger:       logger,
		pollInterval: pollInterval,
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
				// Phase 5 will add due-board lookup and reset execution here.
			}
		}
	}()
}
