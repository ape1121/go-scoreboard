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

var errNotImplemented = errors.New("score repository method not implemented")

type queryer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Repository struct {
	db queryer
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{db: pool}
}

func (r *Repository) Upsert(ctx context.Context, entry score.ScoreEntry) error {
	_, err := r.db.Exec(
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
				achieved_at = EXCLUDED.achieved_at
		`,
		entry.BoardID,
		entry.PeriodID,
		entry.UserID,
		entry.Score,
		entry.AchievedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert score: %w", err)
	}

	return nil
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

func (r *Repository) Surroundings(context.Context, string, int64, string, int) ([]score.ScoreEntry, []score.ScoreEntry, score.ScoreEntry, error) {
	_ = r.db
	return nil, nil, score.ScoreEntry{}, errNotImplemented
}
