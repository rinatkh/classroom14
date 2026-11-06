SHELL := /bin/bash
GO ?= go
BASE_URL ?= http://localhost:8080
HTTP_ADDR ?= :8080
TOKEN_TTL ?= 2h
PG_DUMP ?= pg_dump
export BASE_URL HTTP_ADDR TOKEN_TTL JWT_SECRET DATABASE_URL TEST_DATABASE_URL PG_DUMP

.PHONY: help test-db test-db-race ci-db up down status structure db-up db-down sql-demo run build test test-unit test-race fmt fmt-check vet deps ci curl-demo clean
help:
	@echo "Task API: make run / make curl-demo / make ci"
	@echo "Миграции goose: make up / make down / make status"
	@echo "Снимок полной схемы: structure.sql; обновить: make structure"
run:
	$(GO) run ./cmd/api
build:
	@mkdir -p bin
	$(GO) build -buildvcs=false -o bin/task-api ./cmd/api
	$(GO) build -buildvcs=false -o bin/task-migrate ./cmd/migrate
test:
	$(GO) test ./...
test-unit:
	$(GO) test ./internal/...
test-race:
	$(GO) test -race ./...
fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './bin/*')
fmt-check:
	@files="$$(gofmt -l $$(find . -type f -name '*.go' -not -path './bin/*'))"; \
	if [ -n "$$files" ]; then echo "$$files"; exit 1; fi
vet:
	$(GO) vet ./...
deps:
	$(GO) list -buildvcs=false -deps ./... >/dev/null
ci: fmt-check vet deps test test-race build
curl-demo:
	bash ./scripts/curl-demo.sh
clean:
	rm -f bin/task-api bin/task-migrate coverage.out

# Integration tests create isolated schemas and apply the same migrations.
test-db:
	@test -n "$$TEST_DATABASE_URL" || { echo "Задайте TEST_DATABASE_URL для отдельной тестовой БД"; exit 1; }
	$(GO) test -tags=integration -count=1 ./internal/repository/postgres
test-db-race:
	@test -n "$$TEST_DATABASE_URL" || { echo "Задайте TEST_DATABASE_URL"; exit 1; }
	$(GO) test -race -tags=integration -count=1 ./internal/repository/postgres
ci-db: ci test-db test-db-race
up:
	@test -n "$$DATABASE_URL" || { echo "Задайте DATABASE_URL"; exit 1; }
	$(GO) run ./cmd/migrate up
down:
	@test -n "$$DATABASE_URL" || { echo "Задайте DATABASE_URL"; exit 1; }
	$(GO) run ./cmd/migrate down
status:
	@test -n "$$DATABASE_URL" || { echo "Задайте DATABASE_URL"; exit 1; }
	$(GO) run ./cmd/migrate status
structure:
	bash ./scripts/dump-structure.sh
# Совместимость с прежними командами; рекомендуемые имена — up/down.
db-up: up
db-down: down
sql-demo:
	@test -n "$$DATABASE_URL" || { echo "Задайте DATABASE_URL"; exit 1; }
	psql "$$DATABASE_URL" -X -v ON_ERROR_STOP=1 -f sql/09_lesson_demo.sql
