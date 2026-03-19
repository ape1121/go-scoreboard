package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ape1121/go-scoreboard/internal/board"
)

var errNotImplemented = errors.New("board repository method not implemented")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(context.Context, board.Board, board.BoardPeriod) error {
	_ = r.pool
	return errNotImplemented
}

func (r *Repository) List(context.Context) ([]board.Board, error) {
	_ = r.pool
	return nil, errNotImplemented
}

func (r *Repository) GetByID(context.Context, string) (board.Board, error) {
	_ = r.pool
	return board.Board{}, errNotImplemented
}

func (r *Repository) GetActivePeriod(context.Context, string) (board.BoardPeriod, error) {
	_ = r.pool
	return board.BoardPeriod{}, errNotImplemented
}
