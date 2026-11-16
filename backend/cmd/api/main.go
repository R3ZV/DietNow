// Command api runs the DietNow HTTP API.
//
//	go run ./cmd/api
//
// It serves JSON under /api (see docs/ARCHITECTURE.md). If PostgreSQL is not
// reachable at startup the server still starts, so the liveness endpoint works,
// and GET /api/health/ready reports the database as unavailable.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dietnow/backend/internal/config"
	"github.com/dietnow/backend/internal/database"
	"github.com/dietnow/backend/internal/httpapi"
)

// version is overridden at build time: go build -ldflags "-X main.version=1.2.3"
var version = "dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("api exited with error", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	deps := httpapi.Deps{
		Logger:         logger,
		Version:        version,
		AllowedOrigins: cfg.AllowedOrigins,
	}

	db, err := database.Connect(database.ConfigFromEnv())
	if err != nil {
		logger.Warn("database unavailable; starting without it (readiness will report 503)", "err", err)
	} else {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		defer sqlDB.Close()
		deps.DB = sqlDB
	}

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           httpapi.NewRouter(deps),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api listening", "addr", "http://localhost"+cfg.Addr(), "env", cfg.Env, "version", version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
