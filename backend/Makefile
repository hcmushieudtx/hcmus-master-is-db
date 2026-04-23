BINARY_NAME=bookstore-api
BUILD_DIR=./bin
MIGRATE=migrate
SQLC=sqlc

POSTGRES_DSN?=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATIONS_DIR=db/postgres/migrations

.PHONY: all build run clean migrate-up migrate-down migrate-create sqlc-generate tidy

all: build

## Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

## Run the server (reads .env automatically)
run:
	go run ./main.go server

## Run with live reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

## Tidy and download all Go module dependencies
tidy:
	go mod tidy

## Apply all pending PostgreSQL migrations
migrate-up:
	$(MIGRATE) -source file://$(MIGRATIONS_DIR) -database "$(POSTGRES_DSN)" up

## Roll back one PostgreSQL migration
migrate-down:
	$(MIGRATE) -source file://$(MIGRATIONS_DIR) -database "$(POSTGRES_DSN)" down 1

## Create a new migration pair: make migrate-create NAME=add_something
migrate-create:
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

## Generate sqlc typed query code from db/postgres/queries/
sqlc-generate:
	cd db/postgres && $(SQLC) generate

## Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
