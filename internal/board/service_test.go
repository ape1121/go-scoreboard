package board

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ape1121/go-scoreboard/internal/platform/clock"
)

func TestServiceCreatePersistsBoardAndInitialPeriod(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	repository := &stubRepository{}
	service := NewService(repository, fixedClock{now: fixedNow}, func() string { return "board_test" })

	entity, err := service.Create(context.Background(), CreateInput{
		Name:        " Weekly Tournament ",
		Description: "Global leaderboard",
		Schedule: &Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 7 * 24 * time.Hour,
		},
	})

	require.NoError(t, err)
	require.Equal(t, "board_test", entity.ID)
	require.Equal(t, "Weekly Tournament", entity.Name)
	require.Equal(t, fixedNow, entity.CreatedAt)
	require.NotNil(t, entity.Schedule)
	require.Equal(t, int64(604800), entity.Schedule.IntervalSeconds())
	require.Equal(t, entity, repository.createdBoard)
	require.Equal(t, BoardPeriod{
		BoardID:   "board_test",
		Sequence:  0,
		StartedAt: fixedNow,
	}, repository.createdPeriod)
}

func TestServiceCreateReturnsValidationError(t *testing.T) {
	t.Parallel()

	service := NewService(&stubRepository{}, fixedClock{now: time.Now().UTC()}, func() string { return "board_test" })

	_, err := service.Create(context.Background(), CreateInput{Name: ""})

	var validationErr ValidationError
	require.Error(t, err)
	require.ErrorAs(t, err, &validationErr)
}

func TestServiceGetComputesNextResetAt(t *testing.T) {
	t.Parallel()

	boardEntity := Board{
		ID:          "board_test",
		Name:        "Weekly Tournament",
		Description: "Global leaderboard",
		Schedule: &Schedule{
			Type:     ScheduleTypeInterval,
			Interval: 7 * 24 * time.Hour,
		},
		CreatedAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
	}
	activePeriod := BoardPeriod{
		ID:        1,
		BoardID:   "board_test",
		Sequence:  0,
		StartedAt: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
	}
	repository := &stubRepository{
		boardByID:    boardEntity,
		activePeriod: activePeriod,
	}
	service := NewService(repository, fixedClock{now: time.Now().UTC()}, func() string { return "board_test" })

	details, err := service.Get(context.Background(), "board_test")

	require.NoError(t, err)
	require.Equal(t, boardEntity, details.Board)
	require.NotNil(t, details.NextResetAt)
	require.Equal(t, activePeriod.StartedAt.Add(7*24*time.Hour), *details.NextResetAt)
}

func TestServiceGetWithoutScheduleOmitsNextResetAt(t *testing.T) {
	t.Parallel()

	boardEntity := Board{
		ID:        "board_test",
		Name:      "All Time",
		CreatedAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
	}
	repository := &stubRepository{boardByID: boardEntity}
	service := NewService(repository, fixedClock{now: time.Now().UTC()}, func() string { return "board_test" })

	details, err := service.Get(context.Background(), "board_test")

	require.NoError(t, err)
	require.Nil(t, details.NextResetAt)
}

func TestServiceListDelegatesToRepository(t *testing.T) {
	t.Parallel()

	expected := []Board{
		{ID: "board_1", Name: "Board 1"},
		{ID: "board_2", Name: "Board 2"},
	}
	repository := &stubRepository{listBoards: expected}
	service := NewService(repository, fixedClock{now: time.Now().UTC()}, func() string { return "board_test" })

	boards, err := service.List(context.Background(), 50, 0)

	require.NoError(t, err)
	require.Equal(t, expected, boards)
}

func TestServiceGetReturnsValidationErrorForBlankID(t *testing.T) {
	t.Parallel()

	service := NewService(&stubRepository{}, fixedClock{now: time.Now().UTC()}, func() string { return "board_test" })

	_, err := service.Get(context.Background(), "   ")

	var validationErr ValidationError
	require.Error(t, err)
	require.ErrorAs(t, err, &validationErr)
}

type stubRepository struct {
	createdBoard  Board
	createdPeriod BoardPeriod
	listBoards    []Board
	boardByID     Board
	activePeriod  BoardPeriod
	createErr     error
	listErr       error
	getErr        error
	periodErr     error
}

func (s *stubRepository) Create(_ context.Context, boardEntity Board, period BoardPeriod) error {
	if s.createErr != nil {
		return s.createErr
	}

	s.createdBoard = boardEntity
	s.createdPeriod = period
	return nil
}

func (s *stubRepository) List(_ context.Context, _, _ int) ([]Board, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}

	return s.listBoards, nil
}

func (s *stubRepository) GetByID(_ context.Context, _ string) (Board, error) {
	if s.getErr != nil {
		return Board{}, s.getErr
	}

	return s.boardByID, nil
}

func (s *stubRepository) GetActivePeriod(_ context.Context, _ string) (BoardPeriod, error) {
	if s.periodErr != nil {
		return BoardPeriod{}, s.periodErr
	}

	return s.activePeriod, nil
}

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time {
	return f.now
}

var _ Repository = (*stubRepository)(nil)
var _ clock.Clock = fixedClock{}
var _ error = ValidationError{}

func TestValidationErrorUnwraps(t *testing.T) {
	t.Parallel()

	base := errors.New("base")
	err := NewValidationError(base)

	require.ErrorIs(t, err, base)
}
