# Ручной restart demo: сначала память, затем PostgreSQL

API слушает localhost:8080. Не запускайте classroom13 и classroom14 на одном порту одновременно.
Используйте отдельную Alice email для этого опыта, не email из curl-demo:
скрипт сам регистрирует других пользователей и удаляет созданную им Task.

## Register → login → create

В терминале клиента:

```bash
BASE_URL=http://localhost:8080
EMAIL="alice-persistence-${RANDOM}@example.test"
BODY="{\"email\":\"$EMAIL\",\"password\":\"secret123\"}"
curl -i -X POST "$BASE_URL/register" -H 'Content-Type: application/json' -d "$BODY"
TOKEN="$(curl -fsS -X POST "$BASE_URL/login" -H 'Content-Type: application/json' -d "$BODY" | python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])')"
CREATED="$(curl -fsS -X POST "$BASE_URL/tasks" -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" -d '{"title":"Survive API restart"}')"
TASK_ID="$(printf '%s' "$CREATED" | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')"
curl -i "$BASE_URL/tasks/$TASK_ID" -H "Authorization: Bearer $TOKEN"
```

Ожидаем register 201, login 200, create 201, get 200. Не выводим token в log.

## В classroom13 — данные пропадают

На API terminal остановите API classroom13, сгенерируйте свежий JWT_SECRET для нового
memory-прогона и снова запустите `make -C "../../13 Занятие/classroom13" run`. В клиентском терминале
повторите login с **тем же BODY**, не новым EMAIL: теперь 401.

Повторно register с BODY даёт 201, затем получите новый TOKEN командой login
выше. GET /tasks → [], GET /tasks/$TASK_ID → 404. Старые User/Task исчезли.
После опыта остановите API classroom13.

## В финале — данные переживают restart

Сначала выполните register/login/create блок на финальном API с PostgreSQL,
используя новый EMAIL. Остановите только Go API и снова `make run`, **не меняя**
DATABASE_URL/JWT_SECRET и не выполняя down.

В клиентском терминале с прежними BODY/TOKEN/TASK_ID:

```bash
curl -i -X POST "$BASE_URL/login" -H 'Content-Type: application/json' -d "$BODY"
curl -i "$BASE_URL/tasks/$TASK_ID" -H "Authorization: Bearer $TOKEN"
```

Теперь login и GET снова 200. Данные остались; тот же ещё действующий token
тоже подходит. Повторная регистрация того же email возвращает 409.

## Bob остаётся посторонним

Получите Bob token через отдельные register/login и вызовите PATCH этой Task:

```bash
BOB_EMAIL="bob-persistence-${RANDOM}@example.test"
BOB_BODY="{\"email\":\"$BOB_EMAIL\",\"password\":\"secret123\"}"
curl -i -X POST "$BASE_URL/register" -H 'Content-Type: application/json' -d "$BOB_BODY"
BOB_TOKEN="$(curl -fsS -X POST "$BASE_URL/login" -H 'Content-Type: application/json' -d "$BOB_BODY" | python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])')"
curl -i -X PATCH "$BASE_URL/tasks/$TASK_ID" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -d '{"done":true}'
```

Ожидаем 403 до и после restart. Get/DELETE чужой Task дают тот же 403.
Для полного автоматического Alice/Bob flow есть неизменённый `make curl-demo`.
