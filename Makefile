GO ?= go

# Команды make автоматически читают локальный .env, если он существует.
# Сам Go не загружает этот файл: при прямом `go run` переменные нужно export-нуть.
-include .env

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/go_course_tasks?sslmode=disable
TEST_DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/go_course_tasks_test?sslmode=disable
BASE_URL ?= http://localhost:8080

.PHONY: run test test-race test-db test-db-up test-db-down test-db-reset fmt fmt-check vet ci ci-db db-up db-down db-reset db-seed sql-demo curl-demo

run:
	DATABASE_URL='$(DATABASE_URL)' $(GO) run ./cmd/api

test:
	env -u TEST_DATABASE_URL $(GO) test ./...

test-race:
	env -u TEST_DATABASE_URL $(GO) test -race ./...

test-db: test-db-reset
	TEST_DATABASE_URL='$(TEST_DATABASE_URL)' $(GO) test -count=1 -run Integration ./internal/task

test-db-up:
	psql '$(TEST_DATABASE_URL)' -v ON_ERROR_STOP=1 -f migrations/001_init.up.sql

test-db-down:
	psql '$(TEST_DATABASE_URL)' -v ON_ERROR_STOP=1 -f migrations/001_init.down.sql

test-db-reset:
	$(MAKE) test-db-down
	$(MAKE) test-db-up

fmt:
	gofmt -w $$(find . -type f -name '*.go')

fmt-check:
	@test -z "$$(gofmt -l $$(find . -type f -name '*.go'))" || (echo "Следующие файлы нужно отформатировать:"; gofmt -l $$(find . -type f -name '*.go'); exit 1)

vet:
	$(GO) vet ./...

ci: fmt-check vet test test-race

ci-db: fmt-check vet test test-race test-db

db-up:
	psql '$(DATABASE_URL)' -v ON_ERROR_STOP=1 -f migrations/001_init.up.sql

db-down:
	psql '$(DATABASE_URL)' -v ON_ERROR_STOP=1 -f migrations/001_init.down.sql

db-reset: db-down db-up db-seed

db-seed:
	psql '$(DATABASE_URL)' -v ON_ERROR_STOP=1 -f sql/02_insert.sql

sql-demo:
	psql '$(DATABASE_URL)' -v ON_ERROR_STOP=1 -f sql/09_lesson_demo.sql

curl-demo:
	BASE_URL='$(BASE_URL)' ./scripts/curl-demo.sh
