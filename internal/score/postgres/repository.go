package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ape1121/go-scoreboard/internal/score"
)

var errNotImplemented = errors.New("score repository method not implemented")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Upsert(context.Context, score.ScoreEntry) error {
	_ = r.pool
	return errNotImplemented
}

func (r *Repository) Top(context.Context, string, int64, int) ([]score.ScoreEntry, error) {
	_ = r.pool
	return nil, errNotImplemented
}

func (r *Repository) Get(context.Context, string, int64, string) (score.ScoreEntry, error) {
	_ = r.pool
	return score.ScoreEntry{}, errNotImplemented
}

func (r *Repository) Surroundings(context.Context, string, int64, string, int) ([]score.ScoreEntry, []score.ScoreEntry, score.ScoreEntry, error) {
	_ = r.pool
	return nil, nil, score.ScoreEntry{}, errNotImplemented
}
