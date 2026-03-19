package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ape1121/go-scoreboard/internal/platform/config"
	platformdb "github.com/ape1121/go-scoreboard/internal/platform/db"
	platformhttp "github.com/ape1121/go-scoreboard/internal/platform/http"
	"github.com/ape1121/go-scoreboard/internal/platform/scheduler"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := platformdb.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	sched := scheduler.New(logger, cfg.SchedulerPollInterval)
	sched.Start(ctx)

	server := platformhttp.NewServer(cfg, logger, platformhttp.NewRouter(platformhttp.Dependencies{
		Logger:        logger,
		HealthService: platformhttp.NewHealthService(pool),
	}))

	go func() {
		logger.Printf("http server listening on %s", cfg.HTTPAddr())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("serve http: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("shutdown http server: %v", err)
	}
}
