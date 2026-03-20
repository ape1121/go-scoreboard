package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) DueBoardIDs(ctx context.Context, now time.Time) ([]string, error) {
	rows, err := r.pool.Query(
		ctx,
		`
			SELECT b.board_id
			FROM boards b
			JOIN board_periods bp
				ON bp.board_id = b.board_id
				AND bp.ended_at IS NULL
			WHERE b.schedule_type = 'interval'
				AND b.schedule_interval_seconds IS NOT NULL
				AND bp.started_at + (b.schedule_interval_seconds * INTERVAL '1 second') <= $1
			ORDER BY b.board_id ASC
		`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("query due boards: %w", err)
	}
	defer rows.Close()

	var boardIDs []string
	for rows.Next() {
		var boardID string
		if err := rows.Scan(&boardID); err != nil {
			return nil, fmt.Errorf("scan due board: %w", err)
		}
		boardIDs = append(boardIDs, boardID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate due boards: %w", err)
	}

	return boardIDs, nil
}

func (r *Repository) ResetDueBoard(ctx context.Context, boardID string, now time.Time) (bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin reset transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		intervalSeconds int64
		periodID        int64
		sequence        int64
		startedAt       time.Time
	)

	err = tx.QueryRow(
		ctx,
		`
			SELECT
				b.schedule_interval_seconds,
				bp.period_id,
				bp.sequence_number,
				bp.started_at
			FROM boards b
			JOIN board_periods bp
				ON bp.board_id = b.board_id
				AND bp.ended_at IS NULL
			WHERE b.board_id = $1
				AND b.schedule_type = 'interval'
				AND b.schedule_interval_seconds IS NOT NULL
			FOR UPDATE OF b, bp
		`,
		boardID,
	).Scan(&intervalSeconds, &periodID, &sequence, &startedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("load board reset state: %w", err)
	}

	nextStartedAt, sequenceIncrement, due := advancePeriod(startedAt, time.Duration(intervalSeconds)*time.Second, now)
	if !due {
		return false, nil
	}

	if _, err := tx.Exec(
		ctx,
		`
			UPDATE board_periods
			SET ended_at = $1
			WHERE period_id = $2 AND ended_at IS NULL
		`,
		nextStartedAt,
		periodID,
	); err != nil {
		return false, fmt.Errorf("close active period: %w", err)
	}

	if _, err := tx.Exec(
		ctx,
		`
			DELETE FROM board_scores
			WHERE board_id = $1 AND board_period_id = $2
		`,
		boardID,
		periodID,
	); err != nil {
		return false, fmt.Errorf("delete old scores: %w", err)
	}

	if _, err := tx.Exec(
		ctx,
		`
			INSERT INTO board_periods (
				board_id,
				sequence_number,
				started_at,
				ended_at
			)
			VALUES ($1, $2, $3, NULL)
		`,
		boardID,
		sequence+sequenceIncrement,
		nextStartedAt,
	); err != nil {
		return false, fmt.Errorf("insert next active period: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit reset transaction: %w", err)
	}

	return true, nil
}

func advancePeriod(startedAt time.Time, interval time.Duration, now time.Time) (time.Time, int64, bool) {
	if interval <= 0 {
		return time.Time{}, 0, false
	}

	elapsed := now.Sub(startedAt)
	if elapsed < interval {
		return time.Time{}, 0, false
	}

	increments := int64(elapsed / interval)
	return startedAt.Add(time.Duration(increments) * interval), increments, true
}
