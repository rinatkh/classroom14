#!/usr/bin/env bash
set -euo pipefail

base_url="${BASE_URL:-http://localhost:8080}"
work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

response_body="$work_dir/body.json"

request() {
  local expected_status="$1"
  local label="$2"
  shift 2

  printf '\n%s\n' "$label"
  local actual_status
  actual_status="$(curl -sS -o "$response_body" -w '%{http_code}' "$@")"
  printf 'HTTP status: %s\n' "$actual_status"
  if [[ -s "$response_body" ]]; then
    cat "$response_body"
    printf '\n'
  else
    printf '[body отсутствует]\n'
  fi

  if [[ "$actual_status" != "$expected_status" ]]; then
    printf 'Ожидался status %s, получен %s.\n' "$expected_status" "$actual_status" >&2
    exit 1
  fi
}

request 200 "1. GET /health" "$base_url/health"

request 201 "2. POST /tasks" -X POST "$base_url/tasks" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Понять PostgreSQL"}'

task_id="$(sed -n 's/.*"id":[[:space:]]*\([0-9][0-9]*\).*/\1/p' "$response_body" | head -n 1)"
if [[ -z "$task_id" ]]; then
  printf 'POST вернул 201, но в JSON не найден числовой id.\n' >&2
  exit 1
fi
printf 'Создана задача с id=%s\n' "$task_id"

request 200 "3. GET /tasks" "$base_url/tasks"
request 200 "4. GET /tasks/$task_id" "$base_url/tasks/$task_id"

request 200 "5. PATCH /tasks/$task_id" -X PATCH "$base_url/tasks/$task_id" \
  -H 'Content-Type: application/json' \
  -d '{"done":true}'

request 200 "6. GET /tasks/$task_id после PATCH" "$base_url/tasks/$task_id"
if ! grep -Eq '"done"[[:space:]]*:[[:space:]]*true' "$response_body"; then
  printf 'После PATCH поле done не стало true.\n' >&2
  exit 1
fi

request 204 "7. DELETE /tasks/$task_id" -X DELETE "$base_url/tasks/$task_id"
if [[ -s "$response_body" ]]; then
  printf 'Ответ 204 не должен содержать body.\n' >&2
  exit 1
fi

request 404 "8. GET удалённой задачи" "$base_url/tasks/$task_id"

request 400 "9. POST с ошибочным JSON" -X POST "$base_url/tasks" \
  -H 'Content-Type: application/json' \
  -d '{"title":'

printf '\nПолный HTTP lifecycle выполнен успешно.\n'
