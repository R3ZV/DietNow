.DEFAULT_GOAL := help

.PHONY: help env setup db backend frontend seed stop test build

help:
	@echo "DietNow local development (run from the repository root)"
	@echo "  make setup     Create missing .env files and install dependencies"
	@echo "  make backend   Start PostgreSQL, wait for readiness, and run the API"
	@echo "  make frontend  Run the frontend in a second terminal"
	@echo "  make seed      Start PostgreSQL and load demo data"
	@echo "  make stop      Stop PostgreSQL, preserving its data"
	@echo "  make test      Run backend and frontend tests"
	@echo "  make build     Build backend packages and the frontend"
	@echo "Open http://localhost:5173. Press Ctrl+C in each app terminal to stop it."

# Preserve existing configuration, including credentials for an existing DB.
env:
	@test -f backend/.env || cp backend/.env.example backend/.env
	@test -f frontend/.env || cp frontend/.env.example frontend/.env

setup: env
	cd backend && go mod download
	cd frontend && npm ci
	@echo "Setup complete. For a new database, set DB_PASSWORD in backend/.env before starting."

db: env
	cd backend && docker compose up -d --wait

backend: db
	cd backend && go run ./cmd/api

frontend: env
	cd frontend && npm run dev

# Optional: seed replaces the demo user's profile and today's demo entries.
seed: db
	cd backend && go run ./cmd/seed

stop:
	cd backend && docker compose down

test: db
	cd backend && DIETNOW_REQUIRE_DB=1 go test ./...
	cd frontend && npm test

build:
	cd backend && go build ./...
	cd frontend && npm run build
