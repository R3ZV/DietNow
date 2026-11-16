.DEFAULT_GOAL := help

.PHONY: help env setup dataset import-dataset db backend frontend seed stop test build

DATASET_NAME := FoodData_Central_foundation_food_json_2026-10-29
DATASET_URL := https://fdc.nal.usda.gov/fdc-datasets/$(DATASET_NAME).zip
DATASET_DIR := dataset

help:
	@echo "DietNow local development (run from the repository root)"
	@echo "  make setup     Install dependencies, download USDA foods, and import into PostgreSQL"
	@echo "  make dataset   Download and extract the USDA Foundation dataset"
	@echo "  make import-dataset  Migrate PostgreSQL and import the downloaded foods"
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
	$(MAKE) import-dataset
	@echo "Setup complete. USDA foods are available in PostgreSQL."

# Download atomically; keep the archive and JSON out of git. Existing files are reused.
dataset: $(DATASET_DIR)/$(DATASET_NAME).json

$(DATASET_DIR)/$(DATASET_NAME).zip:
	mkdir -p $(DATASET_DIR)
	curl --fail --location --retry 3 --output $@.tmp $(DATASET_URL)
	unzip -tq $@.tmp
	mv $@.tmp $@

$(DATASET_DIR)/$(DATASET_NAME).json: $(DATASET_DIR)/$(DATASET_NAME).zip
	unzip -p $< $(DATASET_NAME).json > $@.tmp
	test -s $@.tmp
	mv $@.tmp $@

import-dataset: dataset db
	cd backend && go run ./cmd/import-foods -file ../$(DATASET_DIR)/$(DATASET_NAME).json

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
