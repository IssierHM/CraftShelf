# CraftShelf development commands
#
# Migrations require:
#   $env:GOOSE_DRIVER   = "postgres"
#   $env:GOOSE_DBSTRING = "postgres://craftshelf:craftshelf_dev_pwd@localhost:5432/craftshelf?sslmode=disable"
#
# Project tools come from go.mod's tool directive; always call them as `go tool <name>`.

GO ?= go
GOOSE_DIR ?= migrations

.PHONY: help build run test test-race test-cover fmt fmt-check vet lint tidy generate migrate-up migrate-status migrate-down verify-layers check

help:
	@echo "CraftShelf targets:"
	@echo "  build           go build ./..."
	@echo "  run             go run ./cmd/api"
	@echo "  test            go test ./..."
	@echo "  test-race       go test -race ./..."
	@echo "  test-cover      go test -race -coverprofile=coverage.out ./..."
	@echo "  fmt             gofmt -w ."
	@echo "  fmt-check       fail if any file is not gofmt-ed"
	@echo "  vet             go vet ./..."
	@echo "  lint            go tool golangci-lint run ./..."
	@echo "  tidy            go mod tidy"
	@echo "  generate        go tool sqlc generate && go tool wire ./cmd/api"
	@echo "  migrate-up      go tool goose -dir migrations up"
	@echo "  migrate-status  go tool goose -dir migrations status"
	@echo "  migrate-down    go tool goose -dir migrations down"
	@echo "  verify-layers   check that domain has no framework dependencies"
	@echo "  check           fmt-check + vet + lint + test-race + verify-layers"

build:
	$(GO) build ./...

run:
	$(GO) run ./cmd/api

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-cover:
	$(GO) test -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -func=coverage.out | tail -1

fmt:
	gofmt -w .

fmt-check:
	@files="$$(gofmt -l .)"; if [ -n "$$files" ]; then echo "unformatted files:"; echo "$$files"; exit 1; fi

vet:
	$(GO) vet ./...

lint:
	$(GO) tool golangci-lint run ./...

tidy:
	$(GO) mod tidy

generate:
	$(GO) tool sqlc generate
	$(GO) tool wire ./cmd/api

migrate-up:
	$(GO) tool goose -dir $(GOOSE_DIR) up

migrate-status:
	$(GO) tool goose -dir $(GOOSE_DIR) status

migrate-down:
	$(GO) tool goose -dir $(GOOSE_DIR) down

verify-layers:
	@if $(GO) list -deps ./internal/domain | grep -Eq "gin-gonic|jackc|google/uuid|sqlc|aws|redis"; then echo "domain layer is polluted"; exit 1; else echo "domain layer is clean"; fi

check: fmt-check vet lint test-race verify-layers
