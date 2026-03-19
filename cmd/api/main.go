package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ape1121/go-scoreboard/internal/platform/app"
	"github.com/ape1121/go-scoreboard/internal/platform/config"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, cleanup, err := app.Build(ctx, cfg, logger)
	if err != nil {
		logger.Fatalf("build application: %v", err)
	}
	defer cleanup()

	application.Scheduler.Start(ctx)

	go func() {
		logger.Printf("http server listening on %s", cfg.HTTPAddr())
		if err := application.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("serve http: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := application.Server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("shutdown http server: %v", err)
	}
}
