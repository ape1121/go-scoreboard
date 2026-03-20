package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ape1121/go-scoreboard/internal/board"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, entity board.Board, period board.BoardPeriod) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	scheduleType, scheduleIntervalSeconds := scheduleValues(entity.Schedule)
	if _, err := tx.Exec(
		ctx,
		`
			INSERT INTO boards (
				board_id,
				name,
				description,
				schedule_type,
				schedule_interval_seconds,
				created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
		entity.ID,
		entity.Name,
		entity.Description,
		scheduleType,
		scheduleIntervalSeconds,
		entity.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert board: %w", err)
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
			VALUES ($1, $2, $3, $4)
		`,
		period.BoardID,
		period.Sequence,
		period.StartedAt,
		period.EndedAt,
	); err != nil {
		return fmt.Errorf("insert board period: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]board.Board, error) {
	rows, err := r.pool.Query(
		ctx,
		`
			SELECT board_id, name
			FROM boards
			ORDER BY created_at ASC, board_id ASC
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("query boards: %w", err)
	}
	defer rows.Close()

	var boards []board.Board
	for rows.Next() {
		var entity board.Board
		if err := rows.Scan(&entity.ID, &entity.Name); err != nil {
			return nil, fmt.Errorf("scan board: %w", err)
		}
		boards = append(boards, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate boards: %w", err)
	}

	return boards, nil
}

func (r *Repository) GetByID(ctx context.Context, boardID string) (board.Board, error) {
	var (
		entity                  board.Board
		scheduleType            *string
		scheduleIntervalSeconds *int64
	)
	err := r.pool.QueryRow(
		ctx,
		`
			SELECT
				board_id,
				name,
				description,
				schedule_type,
				schedule_interval_seconds,
				created_at
			FROM boards
			WHERE board_id = $1
		`,
		boardID,
	).Scan(
		&entity.ID,
		&entity.Name,
		&entity.Description,
		&scheduleType,
		&scheduleIntervalSeconds,
		&entity.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return board.Board{}, board.ErrNotFound
		}
		return board.Board{}, fmt.Errorf("get board by id: %w", err)
	}

	entity.Schedule = scanSchedule(scheduleType, scheduleIntervalSeconds)
	return entity, nil
}

func (r *Repository) GetActivePeriod(ctx context.Context, boardID string) (board.BoardPeriod, error) {
	var period board.BoardPeriod

	err := r.pool.QueryRow(
		ctx,
		`
			SELECT period_id, board_id, sequence_number, started_at, ended_at
			FROM board_periods
			WHERE board_id = $1 AND ended_at IS NULL
			LIMIT 1
		`,
		boardID,
	).Scan(&period.ID, &period.BoardID, &period.Sequence, &period.StartedAt, &period.EndedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return board.BoardPeriod{}, board.ErrActivePeriodNotFound
		}
		return board.BoardPeriod{}, fmt.Errorf("get active board period: %w", err)
	}

	return period, nil
}

func scheduleValues(schedule *board.Schedule) (*string, *int64) {
	if schedule == nil {
		return nil, nil
	}

	typeValue := string(schedule.Type)
	intervalSeconds := schedule.IntervalSeconds()
	return &typeValue, &intervalSeconds
}

func scanSchedule(typeValue *string, intervalSeconds *int64) *board.Schedule {
	if typeValue == nil || intervalSeconds == nil {
		return nil
	}

	return &board.Schedule{
		Type:     board.ScheduleType(*typeValue),
		Interval: time.Duration(*intervalSeconds) * time.Second,
	}
}
