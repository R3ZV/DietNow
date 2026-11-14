// Command api runs the DietNow HTTP API.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/database"
	"github.com/dietnow/backend/internal/httpapi"
)

func main() {
	database.LoadDotEnv()

	appEnv := getenv("APP_ENV", "local")
	production := appEnv != "local"

	dbCfg := database.ConfigFromEnv()
	if production && (dbCfg.Password == "" || dbCfg.Password == "dietnow_dev_pw") {
		log.Fatal("config: set a real DB_PASSWORD outside local development")
	}
	authCfg := auth.ConfigFromEnv()
	if err := authCfg.Validate(production); err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	authSvc, err := auth.NewService(db, authCfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go purgeExpiredSessions(ctx, authSvc)

	srv := &http.Server{
		Addr: getenv("HTTP_ADDR", ":8080"),
		Handler: httpapi.New(db, authSvc, httpapi.Options{
			AllowedOrigins: strings.Split(getenv("ALLOWED_ORIGINS", ""), ","),
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("DietNow API (%s) listening on %s", appEnv, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func purgeExpiredSessions(ctx context.Context, svc *auth.Service) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		if n, err := svc.DeleteExpired(ctx); err != nil && ctx.Err() == nil {
			log.Printf("purge sessions: %v", err)
		} else if n > 0 {
			log.Printf("purged %d expired sessions", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
