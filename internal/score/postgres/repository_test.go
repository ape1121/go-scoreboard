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

	db := &storeStub{
		tx: &txStub{
			queryRows: []pgx.Row{
				rowStub{values: []any{"board_test"}},
				rowStub{values: []any{int64(11)}},
			},
		},
	}
	repository := &Repository{db: db}
	now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	entry, err := repository.Upsert(context.Background(), score.UpsertInput{
		BoardID:    "board_test",
		UserID:     "user_1",
		Score:      1500,
		AchievedAt: now,
	})

	require.NoError(t, err)
	require.Contains(t, db.tx.querySQL[0], "SELECT board_id FROM boards WHERE board_id = $1 FOR UPDATE")
	require.Contains(t, db.tx.querySQL[1], "SELECT period_id FROM board_periods WHERE board_id = $1 AND ended_at IS NULL LIMIT 1")
	require.Contains(t, db.tx.execSQL, "ON CONFLICT (board_id, board_period_id, user_id)")
	require.Contains(t, db.tx.execSQL, "score = EXCLUDED.score")
	require.Contains(t, db.tx.execSQL, "achieved_at = EXCLUDED.achieved_at")
	require.Equal(t, []any{"board_test", int64(11), "user_1", int64(1500), now}, db.tx.execArgs)
	require.Equal(t, score.ScoreEntry{
		BoardID:    "board_test",
		PeriodID:   11,
		UserID:     "user_1",
		Score:      1500,
		AchievedAt: now,
	}, entry)
}

func TestRepositoryTopUsesRankingQueryAndReturnsEntries(t *testing.T) {
	t.Parallel()

	firstTime := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	secondTime := firstTime.Add(time.Minute)
	db := &storeStub{
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
	querySQL  string
	queryArgs []any
	rows      pgx.Rows
	row       pgx.Row
	queryErr  error
}

type storeStub struct {
	tx        *txStub
	querySQL  string
	queryArgs []any
	rows      pgx.Rows
	row       pgx.Row
	queryErr  error
	beginErr  error
}

func (s *storeStub) BeginTx(context.Context, pgx.TxOptions) (tx, error) {
	if s.beginErr != nil {
		return nil, s.beginErr
	}

	if s.tx == nil {
		s.tx = &txStub{}
	}

	return s.tx, nil
}

func (s *storeStub) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
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

func (s *storeStub) QueryRow(context.Context, string, ...any) pgx.Row {
	if s.row == nil {
		return rowStub{err: errors.New("query row not configured")}
	}

	return s.row
}

type txStub struct {
	querySQL  []string
	execSQL   string
	execArgs  []any
	queryRows []pgx.Row
	execErr   error
	commitErr error
}

func (s *txStub) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	s.execSQL = normalizeWhitespace(sql)
	s.execArgs = args
	return pgconn.CommandTag{}, s.execErr
}

func (s *txStub) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	s.querySQL = append(s.querySQL, normalizeWhitespace(sql))
	if len(s.queryRows) == 0 {
		return rowStub{err: errors.New("query row not configured")}
	}

	row := s.queryRows[0]
	s.queryRows = s.queryRows[1:]
	return row
}

func (s *txStub) Commit(context.Context) error {
	return s.commitErr
}

func (s *txStub) Rollback(context.Context) error {
	return nil
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
