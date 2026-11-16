// Command api runs the DietNow HTTP API (and, when STATIC_DIR is set, serves
// the built frontend from the same origin).
//
//	go run ./cmd/api
//
// It serves JSON under /api (see docs/ARCHITECTURE.md). If PostgreSQL is not
// reachable at startup in development the server still starts, so the
// liveness endpoint works and GET /api/health/ready reports 503. Outside
// development an unreachable database or a failed migration is fatal, so a
// broken release never becomes healthy and the platform keeps the old one.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // IANA zones for profile time zones, even in minimal containers

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/config"
	"github.com/dietnow/backend/internal/database"
	"github.com/dietnow/backend/internal/httpapi"
)

// version is overridden at build time: go build -ldflags "-X main.version=<git sha>"
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
	dbCfg := database.ConfigFromEnv()
	authCfg := auth.ConfigFromEnv()
	if err := authCfg.Validate(cfg.Strict()); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if cfg.Strict() && (dbCfg.Password == "" || dbCfg.Password == "dietnow_dev_pw" || dbCfg.Password == "change-me-local-db-password") {
		return errors.New("config: set a real DB_PASSWORD outside local development")
	}

	deps := httpapi.Deps{
		Logger:         logger,
		Version:        version,
		AllowedOrigins: cfg.AllowedOrigins,
		StaticDir:      cfg.StaticDir,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(dbCfg)
	switch {
	case err != nil && cfg.Strict():
		return fmt.Errorf("connect database: %w", err)
	case err != nil:
		logger.Warn("database unavailable; starting without it (readiness will report 503)", "err", err)
	default:
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		defer sqlDB.Close()

		if cfg.AutoMigrate {
			if err := database.Migrate(db); err != nil {
				return fmt.Errorf("migrate: %w", err)
			}
			if err := database.SeedFoodCatalog(db); err != nil {
				return fmt.Errorf("seed food catalog: %w", err)
			}
			logger.Info("database migrated")
		}

		authSvc, err := auth.NewService(db, authCfg)
		if err != nil {
			return err
		}
		deps.DB, deps.Auth = db, authSvc
		go purgeExpiredSessions(ctx, logger, authSvc)
	}

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           httpapi.NewRouter(deps),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api listening", "addr", cfg.Addr(), "env", cfg.Env, "version", version, "static", cfg.StaticDir != "")
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

func purgeExpiredSessions(ctx context.Context, logger *slog.Logger, svc *auth.Service) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		if n, err := svc.DeleteExpired(ctx); err != nil && ctx.Err() == nil {
			logger.Warn("purge sessions", "err", err)
		} else if n > 0 {
			logger.Info("purged expired sessions", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
