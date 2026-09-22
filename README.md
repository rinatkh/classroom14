# Занятие 15. PostgreSQL: первые данные Backend API

На прошлом backend-занятии Task API хранил задачи в `map`. После остановки
сервера всё исчезало. На этом занятии тот же HTTP API подключается к PostgreSQL:

```text
HTTP request → handler → PostgresStore → SQL → PostgreSQL
                                            ↓
HTTP response ← JSON ← Task ← RETURNING / SELECT
```

Главная цель — понять путь данных, а не выучить SQL наизусть.

## Что разберём

- база данных, схема, таблица, строка и колонка;
- типы `bigint`, `text`, `boolean`, `timestamptz` и значение `NULL`;
- `PRIMARY KEY`, `UNIQUE`, `FOREIGN KEY`, `CHECK`, `NOT NULL`, `DEFAULT`;
- `CREATE`, `INSERT`, `SELECT`, `UPDATE`, `DELETE`, `JOIN`;
- параметры `$1`, `$2` вместо склеивания SQL-строк;
- индексы и чтение простого плана через `EXPLAIN`;
- транзакции `BEGIN`, `COMMIT`, `ROLLBACK`;
- комментарии SQL и `COMMENT ON`;
- trigger, который обновляет `updated_at`;
- подключение Go через `pgxpool`;
- `Query`, `QueryRow`, `Scan`, `Exec` и `pgx.ErrNoRows`;
- HTTP-тесты с fake repository и интеграционный тест с PostgreSQL.

## Структура проекта

```text
classroom/
├── cmd/api/main.go                 запуск сервера и подключение к БД
├── internal/httpapi/               HTTP router и handlers
├── internal/task/                  модель, Repository и PostgresStore
├── migrations/                     создание и удаление схемы
├── sql/                            маленькие SQL-эксперименты урока
├── scripts/curl-demo.sh            ручная проверка API
├── docs/                           установка, памятки и антипримеры
├── Makefile
└── go.mod
```

Разделение небольшое: HTTP-слой знает язык HTTP, `PostgresStore` знает SQL,
а `main` соединяет эти части. Сложную архитектуру в первом уроке по базе не
добавляем.

## Подготовка PostgreSQL

Нужны локальные `postgres` и `psql`. Подробности для macOS, Linux и Windows:
[docs/setup.md](docs/setup.md).

Создайте две базы:

```bash
createdb -U postgres go_course_tasks
createdb -U postgres go_course_tasks_test
```

Если у пользователя `postgres` установлен пароль `postgres`, готовые строки
подключения уже находятся в `.env.example`. При другом пароле измените URL
только у себя и не добавляйте его в Git.

`make` автоматически читает локальный `.env`. Сам Go этот файл не загружает:
для прямого `go run ./cmd/api` нужно сначала выполнить `export DATABASE_URL=...`.
В маршруте занятия используем `make run`, поэтому достаточно создать `.env`.

## Первый запуск

```bash
cp .env.example .env
make db-up
make run
```

В другом терминале:

```bash
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Понять PostgreSQL"}'
curl -i http://localhost:8080/tasks
```

Ожидаем `200`, затем `201 Created`, затем список с сохранённой задачей.
Остановите и снова запустите Go-сервер: задача останется в PostgreSQL.

Готовый сценарий `make curl-demo` проходит весь lifecycle: health, создание,
список, чтение одной задачи, PATCH, повторное чтение, DELETE, `404` и ошибочный
JSON. Скрипт берёт ID из ответа POST, поэтому не предполагает, что он равен `1`.

## Маршрут live coding

1. Открыть `migrations/001_init.up.sql` и прочитать таблицу сверху вниз.
2. Выполнить `make db-up`, затем открыть `psql "$DATABASE_URL"`.
3. Выполнить файлы `sql/01_...` — `sql/08_...` по порядку.
4. На каждом `UPDATE` и `DELETE` сначала показать `WHERE`.
5. В `EXPLAIN` не требовать Index Scan на трёх строках: выбор Seq Scan нормален.
6. Открыть `cmd/api/main.go`: `New` → `Ping` → `defer Close`.
7. Открыть `internal/task/postgres.go` и вместе реализовать запросы.
8. Запустить API и пройти `POST → GET → PATCH → DELETE`.
9. Перезапустить сервер и обсудить, почему данные не исчезли.
10. Запустить unit-тесты, затем интеграционный тест с отдельной БД.

Полный сценарий преподавателя: [docs/live_coding.md](docs/live_coding.md).

## Команды

| Команда | Что делает |
|---|---|
| `make run` | запускает API на `:8080` |
| `make db-up` | применяет первую миграцию |
| `make db-down` | удаляет учебную схему |
| `make db-reset` | пересоздаёт схему и добавляет примеры |
| `make sql-demo` | выполняет полный SQL-сценарий |
| `make test` | запускает быстрые тесты без настоящей БД |
| `make test-db-up` | применяет миграцию к отдельной тестовой БД |
| `make test-db-down` | удаляет учебную схему из тестовой БД |
| `make test-db-reset` | пересоздаёт схему только в тестовой БД |
| `make test-db` | пересоздаёт тестовую схему и запускает интеграционные тесты |
| `make test-race` | проверяет Go-код race detector-ом |
| `make ci` | форматирование, vet, unit- и race-тесты |
| `make ci-db` | то же плюс интеграционный тест |
| `make curl-demo` | проверяет полный HTTP lifecycle уже запущенного API |

`make test-db` безопасно использует `TEST_DATABASE_URL`, а не `DATABASE_URL`.
Команда каждый раз применяет `down/up`, поэтому на чистой тестовой базе нет
скрытого шага с ручным запуском миграции.

## Что важно заметить

- `pgxpool.New` создаёт pool, а `Ping` проверяет реальное соединение.
- `r.Context()` передаётся в каждый запрос к БД. Отмена HTTP-запроса сможет
  остановить ненужную работу с PostgreSQL.
- `QueryRow(...).Scan(...)` получает одну строку; `Query` возвращает несколько.
- `pgx.ErrNoRows` — нормальный результат поиска отсутствующего ID, поэтому API
  переводит его в `404`, а не в `500`.
- SQL-параметры `$1`, `$2` отделяют данные пользователя от текста запроса.
- `RETURNING` позволяет одним запросом изменить строку и получить результат.

## Как не нужно

Антипримеры разобраны в [docs/bad_examples.md](docs/bad_examples.md). Самые
важные: `UPDATE` без `WHERE`, склеивание SQL со вводом пользователя, забытый
`rows.Close`, хранение пароля в Git и индекс «на всякий случай».

## Вопросы ученице

1. Кто теперь выдаёт `id`: Go-код или PostgreSQL?
2. Чем `PRIMARY KEY` отличается от `FOREIGN KEY`?
3. Почему `NULL` — не то же самое, что пустая строка или ноль?
4. Почему `$1` безопаснее `"..." + title`?
5. Когда использовать `QueryRow`, а когда `Query`?
6. Почему отсутствие строки превращается в `404`?
7. Почему небольшой `SELECT` может выбрать `Seq Scan`, даже если индекс есть?
8. Что сохранит `COMMIT` и что отменит `ROLLBACK`?
9. Когда срабатывает trigger `tasks_set_updated_at`?
10. Почему HTTP handler не содержит SQL?
