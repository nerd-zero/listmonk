\
-include .env
export

GOBIN := $(shell go env GOPATH)/bin
DATABASE_URL ?= postgres://listnun:listnun@localhost:5432/listnun?sslmode=disable

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## --- dev environment (docker-compose.yml: control-plane Postgres + Zitadel) ---

.PHONY: dev-up
dev-up: ## Start the local dev stack (Postgres + Zitadel)
	docker compose up -d

.PHONY: dev-down
dev-down: ## Stop the local dev stack
	docker compose down

.PHONY: dev-logs
dev-logs: ## Tail the local dev stack's logs
	docker compose logs -f

.PHONY: setup-zitadel
setup-zitadel: ## Provision the "listnun" OIDC client in the dev Zitadel (prints values for .env)
	web/scripts/setup-zitadel.sh

## --- backend (Go) ---

.PHONY: tools
tools: ## Install swag/migrate CLIs to $(GOBIN) (one-time)
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: build
build: ## go build ./...
	go build ./...

.PHONY: vet
vet: ## go vet ./...
	go vet ./...

.PHONY: fmt
fmt: ## gofmt -w every Go file
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Fail if any Go file isn't gofmt'd
	@test -z "$$(gofmt -l .)" || (echo "not gofmt'd:"; gofmt -l .; exit 1)

.PHONY: test
test: ## go test ./...
	go test ./...

.PHONY: check
check: fmt-check vet build test ## fmt-check + vet + build + test

.PHONY: migrate
migrate: ## Apply pending migrations (go run ./cmd/migrate)
	go run ./cmd/migrate

.PHONY: migrate-down
migrate-down: ## Roll back one migration
	go run ./cmd/migrate -direction down

.PHONY: api
api: ## Run the API server (go run ./cmd/api)
	go run ./cmd/api

.PHONY: worker
worker: ## Run the background worker (go run ./cmd/worker)
	go run ./cmd/worker

.PHONY: swagger
swagger: ## Regenerate the OpenAPI spec (internal/apidocs) from handler doc comments
	@PATH="$(PATH):$(GOBIN)" swag init -g cmd/api/main.go --parseInternal -o internal/apidocs

## --- frontend (web/) ---

.PHONY: web-install
web-install: ## bun install in web/
	cd web && bun install

.PHONY: web-dev
web-dev: ## Run the frontend dev server
	cd web && bun run dev

.PHONY: web-build
web-build: ## Build the frontend for production
	cd web && bun run build

.PHONY: orval
orval: ## Regenerate the frontend's typed API client from internal/apidocs/swagger.json
	cd web && npx orval

## --- release ---

.PHONY: version
version: ## Bump VERSION (CalVer YYYY.MM.XXX)
	@scripts/increment-version.sh

.DEFAULT_GOAL := help
