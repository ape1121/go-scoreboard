package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	"github.com/ape1121/go-scoreboard/internal/score"
)

func TestRepositoryUpsertUsesOverwriteQuery(t *testing.T) {
	t.Parallel()

	db := &queryerStub{}
	repository := &Repository{db: db}
	now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	err := repository.Upsert(context.Background(), score.ScoreEntry{
		BoardID:    "board_test",
		PeriodID:   11,
		UserID:     "user_1",
		Score:      1500,
		AchievedAt: now,
	})

	require.NoError(t, err)
	require.Contains(t, db.execSQL, "ON CONFLICT (board_id, board_period_id, user_id)")
	require.Contains(t, db.execSQL, "score = EXCLUDED.score")
	require.Contains(t, db.execSQL, "achieved_at = EXCLUDED.achieved_at")
	require.Equal(t, []any{"board_test", int64(11), "user_1", int64(1500), now}, db.execArgs)
}

func TestRepositoryTopUsesRankingQueryAndReturnsEntries(t *testing.T) {
	t.Parallel()

	firstTime := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	secondTime := firstTime.Add(time.Minute)
	db := &queryerStub{
		rows: &rowsStub{
			records: []rowRecord{
				{values: []any{"board_test", int64(11), "user_2", int64(2000), firstTime}},
				{values: []any{"board_test", int64(11), "user_1", int64(2000), secondTime}},
			},
		},
	}
	repository := &Repository{db: db}

	entries, err := repository.Top(context.Background(), "board_test", 11, 10)

	require.NoError(t, err)
	require.Contains(t, normalizeWhitespace(db.querySQL), "ORDER BY score DESC, achieved_at ASC, user_id ASC")
	require.Equal(t, []any{"board_test", int64(11), 10}, db.queryArgs)
	require.Equal(t, []score.ScoreEntry{
		{BoardID: "board_test", PeriodID: 11, UserID: "user_2", Score: 2000, AchievedAt: firstTime},
		{BoardID: "board_test", PeriodID: 11, UserID: "user_1", Score: 2000, AchievedAt: secondTime},
	}, entries)
}

type queryerStub struct {
	execSQL   string
	execArgs  []any
	querySQL  string
	queryArgs []any
	rows      pgx.Rows
	row       pgx.Row
	execErr   error
	queryErr  error
}

func (s *queryerStub) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	s.execSQL = normalizeWhitespace(sql)
	s.execArgs = args
	return pgconn.CommandTag{}, s.execErr
}

func (s *queryerStub) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
	s.querySQL = normalizeWhitespace(sql)
	s.queryArgs = args
	if s.queryErr != nil {
		return nil, s.queryErr
	}

	if s.rows == nil {
		return &rowsStub{}, nil
	}

	return s.rows, nil
}

func (s *queryerStub) QueryRow(context.Context, string, ...any) pgx.Row {
	if s.row == nil {
		return rowStub{err: errors.New("query row not configured")}
	}

	return s.row
}

type rowsStub struct {
	index   int
	records []rowRecord
	err     error
}

func (s *rowsStub) Close() {}

func (s *rowsStub) Err() error {
	return s.err
}

func (s *rowsStub) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (s *rowsStub) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (s *rowsStub) Next() bool {
	return s.index < len(s.records)
}

func (s *rowsStub) Scan(dest ...any) error {
	record := s.records[s.index]
	s.index++

	for i := range dest {
		switch value := dest[i].(type) {
		case *string:
			*value = record.values[i].(string)
		case *int64:
			*value = record.values[i].(int64)
		case *time.Time:
			*value = record.values[i].(time.Time)
		default:
			return errors.New("unsupported scan destination")
		}
	}

	return nil
}

func (s *rowsStub) Values() ([]any, error) {
	return nil, nil
}

func (s *rowsStub) RawValues() [][]byte {
	return nil
}

func (s *rowsStub) Conn() *pgx.Conn {
	return nil
}

type rowStub struct {
	values []any
	err    error
}

func (s rowStub) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}

	for i := range dest {
		switch value := dest[i].(type) {
		case *string:
			*value = s.values[i].(string)
		case *int64:
			*value = s.values[i].(int64)
		case *time.Time:
			*value = s.values[i].(time.Time)
		default:
			return errors.New("unsupported scan destination")
		}
	}

	return nil
}

type rowRecord struct {
	values []any
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
