package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ape1121/go-scoreboard/internal/score"
)

type tx interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Commit(context.Context) error
	Rollback(context.Context) error
}

type store interface {
	BeginTx(context.Context, pgx.TxOptions) (tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type poolStore struct {
	pool *pgxpool.Pool
}

func (s poolStore) BeginTx(ctx context.Context, options pgx.TxOptions) (tx, error) {
	return s.pool.BeginTx(ctx, options)
}

func (s poolStore) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return s.pool.Query(ctx, sql, args...)
}

func (s poolStore) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return s.pool.QueryRow(ctx, sql, args...)
}

type Repository struct {
	db store
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: poolStore{pool: pool}}
}

func (r *Repository) Upsert(ctx context.Context, input score.UpsertInput) (score.ScoreEntry, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return score.ScoreEntry{}, fmt.Errorf("begin score upsert transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var boardID string
	if err := tx.QueryRow(
		ctx,
		`
			SELECT board_id
			FROM boards
			WHERE board_id = $1
			FOR UPDATE
		`,
		input.BoardID,
	).Scan(&boardID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return score.ScoreEntry{}, score.ErrBoardNotFound
		}
		return score.ScoreEntry{}, fmt.Errorf("lock board for score upsert: %w", err)
	}

	var periodID int64
	if err := tx.QueryRow(
		ctx,
		`
			SELECT period_id
			FROM board_periods
			WHERE board_id = $1 AND ended_at IS NULL
			LIMIT 1
		`,
		input.BoardID,
	).Scan(&periodID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return score.ScoreEntry{}, score.ErrActivePeriodNotFound
		}
		return score.ScoreEntry{}, fmt.Errorf("load active period for score upsert: %w", err)
	}

	if _, err := tx.Exec(
		ctx,
		`
			INSERT INTO board_scores (
				board_id,
				board_period_id,
				user_id,
				score,
				achieved_at
			)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (board_id, board_period_id, user_id)
			DO UPDATE SET
				score = EXCLUDED.score,
				achieved_at = CASE
					WHEN board_scores.score <> EXCLUDED.score THEN EXCLUDED.achieved_at
					ELSE board_scores.achieved_at
				END
		`,
		input.BoardID,
		periodID,
		input.UserID,
		input.Score,
		input.AchievedAt,
	); err != nil {
		return score.ScoreEntry{}, fmt.Errorf("upsert score: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return score.ScoreEntry{}, fmt.Errorf("commit score upsert transaction: %w", err)
	}

	return score.ScoreEntry{
		BoardID:    input.BoardID,
		PeriodID:   periodID,
		UserID:     input.UserID,
		Score:      input.Score,
		AchievedAt: input.AchievedAt,
	}, nil
}

func (r *Repository) Top(ctx context.Context, boardID string, periodID int64, limit int) ([]score.ScoreEntry, error) {
	rows, err := r.db.Query(
		ctx,
		`
			SELECT
				board_id,
				board_period_id,
				user_id,
				score,
				achieved_at
			FROM board_scores
			WHERE board_id = $1 AND board_period_id = $2
			ORDER BY score DESC, achieved_at ASC, user_id ASC
			LIMIT $3
		`,
		boardID,
		periodID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query top scores: %w", err)
	}
	defer rows.Close()

	var entries []score.ScoreEntry
	for rows.Next() {
		var entry score.ScoreEntry
		if err := rows.Scan(&entry.BoardID, &entry.PeriodID, &entry.UserID, &entry.Score, &entry.AchievedAt); err != nil {
			return nil, fmt.Errorf("scan top score: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top scores: %w", err)
	}

	return entries, nil
}

func (r *Repository) Get(ctx context.Context, boardID string, periodID int64, userID string) (score.ScoreEntry, error) {
	var entry score.ScoreEntry

	err := r.db.QueryRow(
		ctx,
		`
			SELECT
				board_id,
				board_period_id,
				user_id,
				score,
				achieved_at
			FROM board_scores
			WHERE board_id = $1 AND board_period_id = $2 AND user_id = $3
		`,
		boardID,
		periodID,
		userID,
	).Scan(&entry.BoardID, &entry.PeriodID, &entry.UserID, &entry.Score, &entry.AchievedAt)
	if err != nil {
		return score.ScoreEntry{}, fmt.Errorf("get score: %w", err)
	}

	return entry, nil
}

func (r *Repository) Surroundings(ctx context.Context, boardID string, periodID int64, userID string, n int) ([]score.RankedEntry, error) {
	rows, err := r.db.Query(
		ctx,
		`
			WITH scored AS (
				SELECT 
					board_id, board_period_id, user_id, score, achieved_at,
					ROW_NUMBER() OVER (
						ORDER BY score DESC, achieved_at ASC, user_id ASC
					) AS rank
				FROM board_scores
				WHERE board_id = $1 AND board_period_id = $2
			),
			target AS (
				SELECT * FROM scored WHERE user_id = $3
			)
			SELECT s.board_id, s.board_period_id, s.user_id, s.score, s.achieved_at, s.rank
			FROM scored s
			CROSS JOIN target t
			WHERE s.rank BETWEEN t.rank - $4 AND t.rank + $4
			ORDER BY s.score DESC, s.achieved_at ASC, s.user_id ASC
		`,
		boardID,
		periodID,
		userID,
		n,
	)
	if err != nil {
		return nil, fmt.Errorf("query surroundings: %w", err)
	}
	defer rows.Close()

	var entries []score.RankedEntry
	for rows.Next() {
		var entry score.RankedEntry
		if err := rows.Scan(
			&entry.BoardID, &entry.PeriodID, &entry.UserID,
			&entry.Score, &entry.AchievedAt, &entry.Rank,
		); err != nil {
			return nil, fmt.Errorf("scan surroundings entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate surroundings: %w", err)
	}

	if len(entries) == 0 {
		return nil, score.ErrScoreNotFound
	}

	return entries, nil
}
