package score

import (
	"context"
	"errors"
	"strings"

	"github.com/ape1121/go-scoreboard/internal/board"
	"github.com/ape1121/go-scoreboard/internal/platform/clock"
)

type Repository interface {
	Upsert(context.Context, ScoreEntry) error
	Top(context.Context, string, int64, int) ([]ScoreEntry, error)
	Get(context.Context, string, int64, string) (ScoreEntry, error)
	Surroundings(context.Context, string, int64, string, int) ([]ScoreEntry, []ScoreEntry, ScoreEntry, error)
}

type BoardResolver interface {
	GetByID(context.Context, string) (board.Board, error)
	GetActivePeriod(context.Context, string) (board.BoardPeriod, error)
}

type Service struct {
	repository Repository
	boards     BoardResolver
	clock      clock.Clock
}

func NewService(repository Repository, boards BoardResolver, serviceClock clock.Clock) *Service {
	if serviceClock == nil {
		serviceClock = clock.System{}
	}

	return &Service{repository: repository, boards: boards, clock: serviceClock}
}

func (s *Service) Set(ctx context.Context, input SetInput) (ScoreEntry, error) {
	boardID := strings.TrimSpace(input.BoardID)
	userID := strings.TrimSpace(input.UserID)

	if err := validateBoardID(boardID); err != nil {
		return ScoreEntry{}, err
	}
	if err := NewValidationError(ValidateWrite(userID, input.Score)); err != nil {
		return ScoreEntry{}, err
	}

	if _, err := s.boards.GetByID(ctx, boardID); err != nil {
		return ScoreEntry{}, mapBoardError(err)
	}

	period, err := s.boards.GetActivePeriod(ctx, boardID)
	if err != nil {
		return ScoreEntry{}, mapBoardError(err)
	}

	entry := ScoreEntry{
		BoardID:    boardID,
		PeriodID:   period.ID,
		UserID:     userID,
		Score:      input.Score,
		AchievedAt: s.clock.Now(),
	}

	if err := s.repository.Upsert(ctx, entry); err != nil {
		return ScoreEntry{}, err
	}

	return entry, nil
}

func (s *Service) Top(ctx context.Context, boardID string, limit int) ([]ScoreEntry, error) {
	trimmedBoardID := strings.TrimSpace(boardID)

	if err := validateBoardID(trimmedBoardID); err != nil {
		return nil, err
	}
	if err := NewValidationError(ValidateLimit(limit)); err != nil {
		return nil, err
	}

	if _, err := s.boards.GetByID(ctx, trimmedBoardID); err != nil {
		return nil, mapBoardError(err)
	}

	period, err := s.boards.GetActivePeriod(ctx, trimmedBoardID)
	if err != nil {
		return nil, mapBoardError(err)
	}

	return s.repository.Top(ctx, trimmedBoardID, period.ID, limit)
}

func mapBoardError(err error) error {
	switch {
	case errors.Is(err, board.ErrNotFound):
		return ErrBoardNotFound
	default:
		return err
	}
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
