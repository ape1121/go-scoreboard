package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPPort              = 8080
	defaultSchedulerPollInterval = 10 * time.Second
	defaultShutdownTimeout       = 10 * time.Second
	defaultDatabaseURL           = "postgres://postgres:postgres@localhost:5432/scoreboard?sslmode=disable"
	envHTTPPort                  = "HTTP_PORT"
	envDatabaseURL               = "DATABASE_URL"
	envSchedulerPollInterval     = "SCHEDULER_POLL_INTERVAL"
	envShutdownTimeout           = "SHUTDOWN_TIMEOUT"
)

type Config struct {
	HTTPPort              int
	DatabaseURL           string
	SchedulerPollInterval time.Duration
	ShutdownTimeout       time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:              defaultHTTPPort,
		DatabaseURL:           defaultDatabaseURL,
		SchedulerPollInterval: defaultSchedulerPollInterval,
		ShutdownTimeout:       defaultShutdownTimeout,
	}

	if raw := strings.TrimSpace(os.Getenv(envHTTPPort)); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port <= 0 {
			return Config{}, fmt.Errorf("%s must be a positive integer", envHTTPPort)
		}
		cfg.HTTPPort = port
	}

	if raw := strings.TrimSpace(os.Getenv(envDatabaseURL)); raw != "" {
		cfg.DatabaseURL = raw
	}

	if raw := strings.TrimSpace(os.Getenv(envSchedulerPollInterval)); raw != "" {
		value, err := time.ParseDuration(raw)
		if err != nil || value <= 0 {
			return Config{}, fmt.Errorf("%s must be a positive duration", envSchedulerPollInterval)
		}
		cfg.SchedulerPollInterval = value
	}

	if raw := strings.TrimSpace(os.Getenv(envShutdownTimeout)); raw != "" {
		value, err := time.ParseDuration(raw)
		if err != nil || value <= 0 {
			return Config{}, fmt.Errorf("%s must be a positive duration", envShutdownTimeout)
		}
		cfg.ShutdownTimeout = value
	}

	return cfg, nil
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}
