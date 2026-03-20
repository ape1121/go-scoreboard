//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ape1121/go-scoreboard/internal/board"
	boardpostgres "github.com/ape1121/go-scoreboard/internal/board/postgres"
	"github.com/ape1121/go-scoreboard/internal/platform/testdb"
	"github.com/ape1121/go-scoreboard/internal/score"
	scorepostgres "github.com/ape1121/go-scoreboard/internal/score/postgres"
)

var sharedDB *testdb.TestDB

func TestMain(m *testing.M) {
	ctx := context.Background()
	db, err := testdb.New(ctx)
	if err != nil {
		panic("start test database: " + err.Error())
	}
	sharedDB = db
	defer db.Close(ctx)
	m.Run()
}

func TestCreateAndGetBoard(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	repo := boardpostgres.NewRepository(sharedDB.Pool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	err := repo.Create(ctx, board.Board{
		ID:          "board_integ_1",
		Name:        "Integration Board",
		Description: "Test description",
		Schedule: &board.Schedule{
			Type:     board.ScheduleTypeInterval,
			Interval: time.Hour,
		},
		CreatedAt: now,
	}, board.BoardPeriod{
		BoardID:   "board_integ_1",
		Sequence:  0,
		StartedAt: now,
	})
	require.NoError(t, err)

	entity, err := repo.GetByID(ctx, "board_integ_1")
	require.NoError(t, err)
	require.Equal(t, "board_integ_1", entity.ID)
	require.Equal(t, "Integration Board", entity.Name)
	require.Equal(t, "Test description", entity.Description)
	require.NotNil(t, entity.Schedule)
	require.Equal(t, board.ScheduleTypeInterval, entity.Schedule.Type)
	require.Equal(t, time.Hour, entity.Schedule.Interval)
}

func TestCreateBoardWithoutSchedule(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	repo := boardpostgres.NewRepository(sharedDB.Pool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	err := repo.Create(ctx, board.Board{
		ID:        "board_no_sched",
		Name:      "No Schedule",
		CreatedAt: now,
	}, board.BoardPeriod{
		BoardID:   "board_no_sched",
		Sequence:  0,
		StartedAt: now,
	})
	require.NoError(t, err)

	entity, err := repo.GetByID(ctx, "board_no_sched")
	require.NoError(t, err)
	require.Nil(t, entity.Schedule)
}

func TestGetByIDReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	repo := boardpostgres.NewRepository(sharedDB.Pool)

	_, err := repo.GetByID(ctx, "nonexistent")
	require.ErrorIs(t, err, board.ErrNotFound)
}

func TestListBoardsPagination(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	repo := boardpostgres.NewRepository(sharedDB.Pool)
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 5; i++ {
		err := repo.Create(ctx, board.Board{
			ID:        "list_board_" + string(rune('a'+i)),
			Name:      "Board " + string(rune('A'+i)),
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
		}, board.BoardPeriod{
			BoardID:   "list_board_" + string(rune('a'+i)),
			Sequence:  0,
			StartedAt: base.Add(time.Duration(i) * time.Hour),
		})
		require.NoError(t, err)
	}

	all, err := repo.List(ctx, 100, 0)
	require.NoError(t, err)
	require.Len(t, all, 5)

	page, err := repo.List(ctx, 2, 0)
	require.NoError(t, err)
	require.Len(t, page, 2)
	require.Equal(t, "list_board_a", page[0].ID)
	require.Equal(t, "list_board_b", page[1].ID)

	page2, err := repo.List(ctx, 2, 2)
	require.NoError(t, err)
	require.Len(t, page2, 2)
	require.Equal(t, "list_board_c", page2[0].ID)
	require.Equal(t, "list_board_d", page2[1].ID)
}

func TestGetActivePeriod(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	repo := boardpostgres.NewRepository(sharedDB.Pool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	err := repo.Create(ctx, board.Board{
		ID:        "period_board",
		Name:      "Period Board",
		CreatedAt: now,
	}, board.BoardPeriod{
		BoardID:   "period_board",
		Sequence:  0,
		StartedAt: now,
	})
	require.NoError(t, err)

	period, err := repo.GetActivePeriod(ctx, "period_board")
	require.NoError(t, err)
	require.Equal(t, "period_board", period.BoardID)
	require.EqualValues(t, 0, period.Sequence)
	require.Nil(t, period.EndedAt)
}

func TestResetDueBoardIntegration(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	scoreRepo := scorepostgres.NewRepository(sharedDB.Pool)

	start := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	err := boardRepo.Create(ctx, board.Board{
		ID:   "reset_board",
		Name: "Reset Board",
		Schedule: &board.Schedule{
			Type:     board.ScheduleTypeInterval,
			Interval: time.Hour,
		},
		CreatedAt: start,
	}, board.BoardPeriod{
		BoardID:   "reset_board",
		Sequence:  0,
		StartedAt: start,
	})
	require.NoError(t, err)

	_, err = scoreRepo.Upsert(ctx, score.UpsertInput{
		BoardID:    "reset_board",
		UserID:     "alice",
		Score:      1000,
		AchievedAt: start.Add(time.Minute),
	})
	require.NoError(t, err)

	resetTime := start.Add(2 * time.Hour)
	reset, err := boardRepo.ResetDueBoard(ctx, "reset_board", resetTime)
	require.NoError(t, err)
	require.True(t, reset)

	period, err := boardRepo.GetActivePeriod(ctx, "reset_board")
	require.NoError(t, err)
	require.EqualValues(t, 2, period.Sequence)

	entries, err := scoreRepo.Top(ctx, "reset_board", period.ID, 10)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestDueBoardIDsIntegration(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	repo := boardpostgres.NewRepository(sharedDB.Pool)
	start := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	err := repo.Create(ctx, board.Board{
		ID:   "due_board",
		Name: "Due Board",
		Schedule: &board.Schedule{
			Type:     board.ScheduleTypeInterval,
			Interval: time.Hour,
		},
		CreatedAt: start,
	}, board.BoardPeriod{
		BoardID:   "due_board",
		Sequence:  0,
		StartedAt: start,
	})
	require.NoError(t, err)

	err = repo.Create(ctx, board.Board{
		ID:        "notdue_board",
		Name:      "Not Due Board",
		CreatedAt: start,
	}, board.BoardPeriod{
		BoardID:   "notdue_board",
		Sequence:  0,
		StartedAt: start,
	})
	require.NoError(t, err)

	ids, err := repo.DueBoardIDs(ctx, start.Add(2*time.Hour))
	require.NoError(t, err)
	require.Equal(t, []string{"due_board"}, ids)

	ids, err = repo.DueBoardIDs(ctx, start.Add(30*time.Minute))
	require.NoError(t, err)
	require.Empty(t, ids)
}

func TestMigrationsApplyCleanly(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	repo := boardpostgres.NewRepository(sharedDB.Pool)

	err := repo.Create(ctx, board.Board{
		ID:        "migration_test",
		Name:      "Migration Test",
		CreatedAt: time.Now().UTC(),
	}, board.BoardPeriod{
		BoardID:   "migration_test",
		Sequence:  0,
		StartedAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	entity, err := repo.GetByID(ctx, "migration_test")
	require.NoError(t, err)
	require.Equal(t, "migration_test", entity.ID)
}
