package score

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ape1121/go-scoreboard/internal/board"
	"github.com/ape1121/go-scoreboard/internal/platform/clock"
)

func TestServiceSetUpsertsScoreInActivePeriod(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	repository := &repositoryStub{}
	service := NewService(repository, &boardResolverStub{}, fixedClock{now: now})

	entry, err := service.Set(context.Background(), SetInput{
		BoardID: " board_test ",
		UserID:  " user_1 ",
		Score:   1500,
	})

	require.NoError(t, err)
	require.Equal(t, ScoreEntry{
		BoardID:    "board_test",
		PeriodID:   11,
		UserID:     "user_1",
		Score:      1500,
		AchievedAt: now,
	}, entry)
	require.Equal(t, UpsertInput{
		BoardID:    "board_test",
		UserID:     "user_1",
		Score:      1500,
		AchievedAt: now,
	}, repository.upsertInput)
}

func TestServiceSetReturnsValidationError(t *testing.T) {
	t.Parallel()

	service := NewService(&repositoryStub{}, &boardResolverStub{}, fixedClock{now: time.Now().UTC()})

	_, err := service.Set(context.Background(), SetInput{BoardID: "board_test", UserID: "", Score: 10})

	var validationErr ValidationError
	require.Error(t, err)
	require.ErrorAs(t, err, &validationErr)
}

func TestServiceSetReturnsBoardNotFound(t *testing.T) {
	t.Parallel()

	service := NewService(&repositoryStub{upsertErr: ErrBoardNotFound}, &boardResolverStub{}, fixedClock{now: time.Now().UTC()})

	_, err := service.Set(context.Background(), SetInput{BoardID: "board_test", UserID: "user_1", Score: 10})

	require.ErrorIs(t, err, ErrBoardNotFound)
}

func TestServiceTopReturnsScoresForActivePeriod(t *testing.T) {
	t.Parallel()

	expected := []ScoreEntry{
		{BoardID: "board_test", PeriodID: 11, UserID: "user_1", Score: 1500},
		{BoardID: "board_test", PeriodID: 11, UserID: "user_2", Score: 1400},
	}
	repository := &repositoryStub{topEntries: expected}
	boards := &boardResolverStub{
		boardEntity: board.Board{ID: "board_test"},
		period:      board.BoardPeriod{ID: 11, BoardID: "board_test"},
	}
	service := NewService(repository, boards, fixedClock{now: time.Now().UTC()})

	entries, err := service.Top(context.Background(), "board_test", 10)

	require.NoError(t, err)
	require.Equal(t, expected, entries)
	require.Equal(t, topCall{boardID: "board_test", periodID: 11, limit: 10}, repository.topCall)
}

func TestServiceTopReturnsValidationErrorForInvalidLimit(t *testing.T) {
	t.Parallel()

	service := NewService(&repositoryStub{}, &boardResolverStub{}, fixedClock{now: time.Now().UTC()})

	_, err := service.Top(context.Background(), "board_test", 0)

	var validationErr ValidationError
	require.Error(t, err)
	require.ErrorAs(t, err, &validationErr)
}

func TestValidationErrorUnwraps(t *testing.T) {
	t.Parallel()

	base := errors.New("base")
	err := NewValidationError(base)

	require.ErrorIs(t, err, base)
}

type repositoryStub struct {
	upsertInput UpsertInput
	upserted    ScoreEntry
	topEntries  []ScoreEntry
	upsertErr   error
	topErr      error
	topCall     topCall
}

func (s *repositoryStub) Upsert(_ context.Context, input UpsertInput) (ScoreEntry, error) {
	if s.upsertErr != nil {
		return ScoreEntry{}, s.upsertErr
	}

	s.upsertInput = input
	s.upserted = ScoreEntry{
		BoardID:    input.BoardID,
		PeriodID:   11,
		UserID:     input.UserID,
		Score:      input.Score,
		AchievedAt: input.AchievedAt,
	}
	return s.upserted, nil
}

func (s *repositoryStub) Top(_ context.Context, boardID string, periodID int64, limit int) ([]ScoreEntry, error) {
	if s.topErr != nil {
		return nil, s.topErr
	}

	s.topCall = topCall{boardID: boardID, periodID: periodID, limit: limit}
	return s.topEntries, nil
}

func (s *repositoryStub) Get(context.Context, string, int64, string) (ScoreEntry, error) {
	return ScoreEntry{}, nil
}

func (s *repositoryStub) Surroundings(_ context.Context, _ string, _ int64, _ string, _ int) ([]RankedEntry, error) {
	return nil, nil
}

type boardResolverStub struct {
	boardEntity board.Board
	period      board.BoardPeriod
	getErr      error
	periodErr   error
}

func (s *boardResolverStub) GetByID(context.Context, string) (board.Board, error) {
	if s.getErr != nil {
		return board.Board{}, s.getErr
	}

	return s.boardEntity, nil
}

func (s *boardResolverStub) GetActivePeriod(context.Context, string) (board.BoardPeriod, error) {
	if s.periodErr != nil {
		return board.BoardPeriod{}, s.periodErr
	}

	return s.period, nil
}

type topCall struct {
	boardID  string
	periodID int64
	limit    int
}

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time {
	return f.now
}

var _ Repository = (*repositoryStub)(nil)
var _ BoardResolver = (*boardResolverStub)(nil)
var _ clock.Clock = fixedClock{}
