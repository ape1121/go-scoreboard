package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/ape1121/go-scoreboard/internal/platform/config"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	migrationsPath := "migrations"
	hasFiles, err := hasMigrationFiles(migrationsPath)
	if err != nil {
		logger.Fatalf("inspect migrations: %v", err)
	}
	if !hasFiles {
		logger.Printf("no migration files found in %q", migrationsPath)
		return
	}

	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(filepath.Clean(migrationsPath)))
	migrator, err := migrate.New(sourceURL, cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("create migrator: %v", err)
	}
	defer func() {
		srcErr, dbErr := migrator.Close()
		if srcErr != nil {
			logger.Printf("close migration source: %v", srcErr)
		}
		if dbErr != nil {
			logger.Printf("close migration database: %v", dbErr)
		}
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatalf("apply migrations: %v", err)
	}

	logger.Print("migrations applied")
}

func hasMigrationFiles(root string) (bool, error) {
	var hasFiles bool

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".sql") {
			hasFiles = true
			return fs.SkipAll
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return hasFiles, nil
}
