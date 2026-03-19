package score

import (
	"context"

	"github.com/ape1121/go-scoreboard/internal/platform/clock"
)

type Repository interface {
	Upsert(context.Context, ScoreEntry) error
	Top(context.Context, string, int64, int) ([]ScoreEntry, error)
	Get(context.Context, string, int64, string) (ScoreEntry, error)
	Surroundings(context.Context, string, int64, string, int) ([]ScoreEntry, []ScoreEntry, ScoreEntry, error)
}

type Service struct {
	repository Repository
	clock      clock.Clock
}

func NewService(repository Repository, clock clock.Clock) *Service {
	return &Service{repository: repository, clock: clock}
}
