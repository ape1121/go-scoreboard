//go:build integration

package postgres_test

import (
	"context"
	"sync"
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

func seedBoard(t *testing.T, ctx context.Context, boardID string, schedule *board.Schedule) {
	t.Helper()
	repo := boardpostgres.NewRepository(sharedDB.Pool)
	err := repo.Create(ctx, board.Board{
		ID:        boardID,
		Name:      "Test Board",
		Schedule:  schedule,
		CreatedAt: time.Now().UTC(),
	}, board.BoardPeriod{
		BoardID:   boardID,
		Sequence:  0,
		StartedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
}

func TestUpsertAndTopRankingOrder(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "rank_board", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)

	base := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	_, err := repo.Upsert(ctx, score.UpsertInput{BoardID: "rank_board", UserID: "alice", Score: 1000, AchievedAt: base})
	require.NoError(t, err)
	_, err = repo.Upsert(ctx, score.UpsertInput{BoardID: "rank_board", UserID: "bob", Score: 2000, AchievedAt: base.Add(time.Minute)})
	require.NoError(t, err)
	_, err = repo.Upsert(ctx, score.UpsertInput{BoardID: "rank_board", UserID: "carol", Score: 2000, AchievedAt: base})
	require.NoError(t, err)
	_, err = repo.Upsert(ctx, score.UpsertInput{BoardID: "rank_board", UserID: "dave", Score: 2000, AchievedAt: base})
	require.NoError(t, err)

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "rank_board")
	require.NoError(t, err)

	entries, err := repo.Top(ctx, "rank_board", period.ID, 10)
	require.NoError(t, err)
	require.Len(t, entries, 4)

	require.Equal(t, "carol", entries[0].UserID)
	require.Equal(t, int64(2000), entries[0].Score)

	require.Equal(t, "dave", entries[1].UserID)
	require.Equal(t, int64(2000), entries[1].Score)

	require.Equal(t, "bob", entries[2].UserID)
	require.Equal(t, int64(2000), entries[2].Score)

	require.Equal(t, "alice", entries[3].UserID)
	require.Equal(t, int64(1000), entries[3].Score)
}

func TestUpsertOverwritesSameUser(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "overwrite_board", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)

	base := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	_, err := repo.Upsert(ctx, score.UpsertInput{BoardID: "overwrite_board", UserID: "alice", Score: 100, AchievedAt: base})
	require.NoError(t, err)
	_, err = repo.Upsert(ctx, score.UpsertInput{BoardID: "overwrite_board", UserID: "alice", Score: 999, AchievedAt: base.Add(time.Minute)})
	require.NoError(t, err)

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "overwrite_board")
	require.NoError(t, err)

	entries, err := repo.Top(ctx, "overwrite_board", period.ID, 10)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, int64(999), entries[0].Score)
	require.Equal(t, base.Add(time.Minute).UTC(), entries[0].AchievedAt.UTC())
}

func TestTopReturnsEmptyForBoardWithNoScores(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "empty_board", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "empty_board")
	require.NoError(t, err)

	entries, err := repo.Top(ctx, "empty_board", period.ID, 10)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestTopRespectsLimit(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "limit_board", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)
	base := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 20; i++ {
		_, err := repo.Upsert(ctx, score.UpsertInput{
			BoardID:    "limit_board",
			UserID:     "user_" + string(rune('a'+i)),
			Score:      int64(i * 100),
			AchievedAt: base,
		})
		require.NoError(t, err)
	}

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "limit_board")
	require.NoError(t, err)

	entries, err := repo.Top(ctx, "limit_board", period.ID, 5)
	require.NoError(t, err)
	require.Len(t, entries, 5)
	require.True(t, entries[0].Score >= entries[4].Score)
}

func TestSurroundingsReturnsWindow(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "surr_board", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)
	base := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 10; i++ {
		_, err := repo.Upsert(ctx, score.UpsertInput{
			BoardID:    "surr_board",
			UserID:     "player_" + string(rune('a'+i)),
			Score:      int64((10 - i) * 100),
			AchievedAt: base.Add(time.Duration(i) * time.Minute),
		})
		require.NoError(t, err)
	}

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "surr_board")
	require.NoError(t, err)

	entries, err := repo.Surroundings(ctx, "surr_board", period.ID, "player_e", 2)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	found := false
	for _, e := range entries {
		if e.UserID == "player_e" {
			found = true
		}
	}
	require.True(t, found, "target user should appear in surroundings")

	for i := 1; i < len(entries); i++ {
		require.True(t, entries[i-1].Rank < entries[i].Rank)
	}
}

func TestSurroundingsReturnsErrorForMissingUser(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "surr_404_board", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "surr_404_board")
	require.NoError(t, err)

	_, err = repo.Surroundings(ctx, "surr_404_board", period.ID, "nonexistent", 5)
	require.ErrorIs(t, err, score.ErrScoreNotFound)
}

func TestSurroundingsClampsBorderUsers(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "surr_clamp", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)
	base := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 5; i++ {
		_, err := repo.Upsert(ctx, score.UpsertInput{
			BoardID:    "surr_clamp",
			UserID:     "p_" + string(rune('a'+i)),
			Score:      int64((5 - i) * 100),
			AchievedAt: base,
		})
		require.NoError(t, err)
	}

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "surr_clamp")
	require.NoError(t, err)

	entries, err := repo.Surroundings(ctx, "surr_clamp", period.ID, "p_a", 3)
	require.NoError(t, err)
	require.Equal(t, 1, entries[0].Rank)
	require.Equal(t, "p_a", entries[0].UserID)
}

func TestConcurrentScoreUpserts(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "race_board", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)
	base := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	errCh := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := repo.Upsert(ctx, score.UpsertInput{
				BoardID:    "race_board",
				UserID:     "user_1",
				Score:      int64(idx * 10),
				AchievedAt: base.Add(time.Duration(idx) * time.Millisecond),
			})
			if err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "race_board")
	require.NoError(t, err)

	entries, err := repo.Top(ctx, "race_board", period.ID, 10)
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestConcurrentMultiUserScoreUpserts(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, sharedDB.Truncate(ctx))

	seedBoard(t, ctx, "multi_race_board", nil)
	repo := scorepostgres.NewRepository(sharedDB.Pool)
	base := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go func(user, iter int) {
				defer wg.Done()
				_, err := repo.Upsert(ctx, score.UpsertInput{
					BoardID:    "multi_race_board",
					UserID:     "user_" + string(rune('a'+user)),
					Score:      int64(iter * 100),
					AchievedAt: base.Add(time.Duration(iter) * time.Millisecond),
				})
				if err != nil {
					errCh <- err
				}
			}(i, j)
		}
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}

	boardRepo := boardpostgres.NewRepository(sharedDB.Pool)
	period, err := boardRepo.GetActivePeriod(ctx, "multi_race_board")
	require.NoError(t, err)

	entries, err := repo.Top(ctx, "multi_race_board", period.ID, 20)
	require.NoError(t, err)
	require.Len(t, entries, 10)
}
