package board

import (
	"context"
	"strings"
	"time"

	"github.com/ape1121/go-scoreboard/internal/platform/clock"
)

type Repository interface {
	Create(context.Context, Board, BoardPeriod) error
	List(ctx context.Context, limit, offset int) ([]Board, error)
	GetByID(context.Context, string) (Board, error)
	GetActivePeriod(context.Context, string) (BoardPeriod, error)
}

type Service struct {
	repository Repository
	clock      clock.Clock
	newID      func() string
}

func NewService(repository Repository, serviceClock clock.Clock, newID func() string) *Service {
	if serviceClock == nil {
		serviceClock = clock.System{}
	}
	if newID == nil {
		newID = NewID
	}

	return &Service{repository: repository, clock: serviceClock, newID: newID}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Board, error) {
	schedule := cloneSchedule(input.Schedule)
	name := strings.TrimSpace(input.Name)
	if err := NewValidationError(ValidateNewBoard(name, input.Description, schedule)); err != nil {
		return Board{}, err
	}

	now := s.clock.Now()
	entity := Board{
		ID:          s.newID(),
		Name:        name,
		Description: input.Description,
		Schedule:    schedule,
		CreatedAt:   now,
	}
	period := BoardPeriod{
		BoardID:   entity.ID,
		Sequence:  0,
		StartedAt: now,
	}

	if err := s.repository.Create(ctx, entity, period); err != nil {
		return Board{}, err
	}

	return entity, nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Board, error) {
	return s.repository.List(ctx, limit, offset)
}

func (s *Service) Get(ctx context.Context, boardID string) (Details, error) {
	if err := validateBoardID(boardID); err != nil {
		return Details{}, err
	}

	entity, err := s.repository.GetByID(ctx, strings.TrimSpace(boardID))
	if err != nil {
		return Details{}, err
	}

	if !entity.HasSchedule() {
		return Details{Board: entity}, nil
	}

	period, err := s.repository.GetActivePeriod(ctx, entity.ID)
	if err != nil {
		return Details{}, err
	}

	nextResetAt := period.StartedAt.Add(entity.Schedule.Interval)
	return Details{
		Board:       entity,
		NextResetAt: timePointer(nextResetAt),
	}, nil
}

func cloneSchedule(schedule *Schedule) *Schedule {
	if schedule == nil {
		return nil
	}

	copy := *schedule
	return &copy
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func validateBoardID(boardID string) error {
	if strings.TrimSpace(boardID) == "" {
		return ValidationError{err: errorString("board ID must not be empty")}
	}

	return nil
}

type errorString string

func (e errorString) Error() string {
	return string(e)
}
