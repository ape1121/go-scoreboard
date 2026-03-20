package testdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type TestDB struct {
	Pool      *pgxpool.Pool
	container testcontainers.Container
}

func New(ctx context.Context) (*TestDB, error) {
	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("scoreboard_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	if err := runMigrations(connStr); err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &TestDB{Pool: pool, container: container}, nil
}

func (db *TestDB) Close(ctx context.Context) {
	db.Pool.Close()
	db.container.Terminate(ctx)
}

func (db *TestDB) Truncate(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx, `
		TRUNCATE board_scores, board_periods, boards CASCADE
	`)
	return err
}

func runMigrations(connStr string) error {
	migrationsDir := migrationsPath()
	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(filepath.Clean(migrationsDir)))

	m, err := migrate.New(sourceURL, connStr)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func migrationsPath() string {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename))))
	path := filepath.Join(root, "migrations")

	if _, err := os.Stat(path); err == nil {
		return path
	}

	return "migrations"
}
