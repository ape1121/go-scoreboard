package board

import (
	"context"

	"github.com/ape1121/go-scoreboard/internal/platform/clock"
)

type Repository interface {
	Create(context.Context, Board, BoardPeriod) error
	List(context.Context) ([]Board, error)
	GetByID(context.Context, string) (Board, error)
	GetActivePeriod(context.Context, string) (BoardPeriod, error)
}

type Service struct {
	repository Repository
	clock      clock.Clock
}

func NewService(repository Repository, clock clock.Clock) *Service {
	return &Service{repository: repository, clock: clock}
}
