include .env
export $(shell sed 's/=.*//' .env 2>/dev/null || true)

GOOSE_DRIVER=postgres
MIGRATIONS_DIR=./migrations
GOOSE_DBSTRING=host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) sslmode=disable

.PHONY: migrate-up migrate-down migrate-status run build docker-up docker-down docker-build

migrate-up:
	@echo "Running migrations..."
	@goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down:
	@echo "Rolling back migration..."
	@goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

migrate-status:
	@goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

run:
	@echo "Running application..."
	@go run ./cmd/run

build:
	@echo "Building application..."
	@go build -o bin/app ./cmd/run

docker-up:
	@echo "Starting docker-compose..."
	@docker-compose up -d

docker-down:
	@echo "Stopping docker-compose..."
	@docker-compose down

docker-build:
	@echo "Building docker image..."
	@docker-compose build

docker-up-build: docker-build docker-up

test:
	@echo "Running tests..."
	@go test ./...

