# Classroom 14 — PostgreSQL для того же Task API

Это первое занятие раздела PostgreSQL. Исходник — **финальный Classroom 13**.
Для сравнения открываем его настоящий каталог, без вложенной копии проекта.
Там уже есть архитектура, User, bcrypt, JWT,
authentication middleware, request context, Task.OwnerID и authorization.
Проблема одна: Memory Repositories теряют пользователей и задачи после restart.

В финале заменяем **оба** Repository и wiring. Handler, DTO/Mapper, UseCase,
JWT, middleware, password hashing, ownership и HTTP contract не переписываем.
Так показываем, зачем на занятиях 12–13 вводили Repository interfaces.

## Главный diff

```diff
- userRepository := memory.NewUserRepository()
- taskRepository := memory.NewTaskRepository()
+ // Сначала создать и проверить pool подключения.
+ userRepository := postgres.NewUserRepository(pool)
+ taskRepository := postgres.NewTaskRepository(pool)
```

Ниже остаются те же passwordHasher, tokenManager, authUseCase, taskUseCase,
handlers, authMiddleware и router. Полный diff main и проверка неизменности
остальных файлов описаны в [continuity](docs/continuity.md).

Было: Handler → UseCase → Repository interface ← Memory implementation.
Стало: Handler → UseCase → тот же Repository interface ← Postgres implementation.
Прямого Handler → Postgres пути нет. В task/user нет SQL или типов драйвера.

## Сравнение с исходной версией

Go остаётся **1.22.x** (`go 1.22.0`); нового toolchain урок не требует.

```bash
export JWT_SECRET="$(openssl rand -hex 32)"
make -C "../../13 Занятие/classroom13" run
# Во втором терминале из classroom13:
make curl-demo
```

Для эксперимента register/login/create и потери данных при restart используйте
[шаг 0 live coding](docs/live-coding-script.md). Рабочий старт — classroom13/cmd/api.
При отдельном репозитории укажите свой путь к classroom13. Для работы classroom14
соседний проект не требуется.

## Финальная версия с persistence

Установите PostgreSQL 16+, psql и pg_dump, создайте отдельные локальные учебные базы по
[docs/setup.md](docs/setup.md). Затем в первом терминале:

```bash
export DATABASE_URL='postgres://course_user:course-local-password@localhost:5432/go_course_tasks?sslmode=disable'
export TEST_DATABASE_URL='postgres://course_user:course-local-password@localhost:5432/go_course_tasks_test?sslmode=disable'
export JWT_SECRET="$(openssl rand -hex 32)"
make up
make run
```

Пример credentials относится к локальному setup из инструкции. Makefile не
читает .env автоматически; значения нужно export-нуть. Остановите API classroom13,
если он занял 8080. HTTP_ADDR и TOKEN_TTL работают как в 13, defaults :8080 и 2h.

Во втором терминале:

```bash
make curl-demo
```

Это **тот же curl script из 13**: register/login Alice/Bob, Task CRUD, 401,
403 на чужие операции, partial PATCH и прежние business conflicts. Его сценарий
сам удаляет созданную Task; для persistence demo оставьте отдельную задачу по
[docs/persistence-demo.md](docs/persistence-demo.md).

После restart с тем же DATABASE_URL и JWT_SECRET login снова работает и Task
остаётся. Ключ между перезапусками API не пересоздаём. Пользовательские ID теперь
сохраняются в БД; один и тот же ещё действующий JWT тоже продолжает работать.

## Прежний HTTP contract

| Endpoint | Доступ | Успех |
|---|---|---:|
| GET /health | public, {"status":"ok"} | 200 |
| POST /register | public, User DTO id/email | 201 |
| POST /login | public, token | 200 |
| GET /tasks | authenticated, только свои Task | 200 |
| GET /tasks/{id} | owner | 200 |
| POST /tasks | authenticated, OwnerID из identity | 201 + Location |
| PATCH /tasks/{id} | owner, partial update | 200 |
| DELETE /tasks/{id} | owner, без body | 204 |

Task JSON остаётся id/title/done; OwnerID и даты внутренние. Порядок списка
сохранён: id ASC. Ошибки прежние: 400 для входа, 401 для authentication, 403 для
ownership, 404 для отсутствующей Task, 405 для method, 409 для duplicate email/
повторного завершения/лимита, нейтральный 500 для внутренних ошибок.

## Схема из существующего домена

User 1:N Task. Две таблицы из migration 001:

| Таблица | Поля и ограничения |
|---|---|
| users | id bigint identity PK; email text NOT NULL UNIQUE; password_hash text NOT NULL; created_at timestamptz DEFAULT now() |
| tasks | id bigint identity PK; owner_id NOT NULL FK; title NOT NULL CHECK btrim(title) <> ''; done NOT NULL DEFAULT false; timestamps DEFAULT now() |

FK ссылается на users(id). Выбран ON DELETE CASCADE: удаление учебного User
удаляет его Task. Это осознанная политика данных; HTTP endpoint удаления User
не добавляем. Если требовалось бы хранить историю после удаления, выбрали бы
другую стратегию. Up создаёт users → tasks, down удаляет tasks → users.

Email нормализуется прежним AuthUseCase; extension не добавляем. Repository
сопоставляет unique violation users_email_key с user.ErrAlreadyExists, а
pgx.ErrNoRows — с существующими ErrNotFound. FK/внутренние ошибки остаются за
нейтральным HTTP mapping. Пароль хранится только как прежний bcrypt hash.

Индекс `(owner_id, id)` соответствует настоящему `WHERE owner_id=$1 ORDER BY
id ASC` из урока13. Порядок не меняем на created_at DESC ради другого индекса.
Production Update явно меняет updated_at; trigger оставлен отдельным optional lab.

## Где появился новый код

```text
cmd/api/main.go                      pool + выбор двух postgres implementations
cmd/migrate/main.go                  отдельные команды goose up/down/status
internal/repository/postgres/
    db.go                           только инфраструктурный DBTX
    user.go                         прежний user.Repository через SQL
    task.go                         прежний task.Repository через SQL
    *_integration_test.go           отдельные real PostgreSQL checks
internal/repository/memory/           сохранены reference/fast-test implementations
internal/task, user, auth, httpapi/   прежние Model/UseCase/DTO/Mapper/HTTP/auth
migrations/001_init.sql              goose Up и Down в одном файле
structure.sql                        полный schema-only снимок public
scripts/dump-structure.sh            обновление снимка через pg_dump
sql/                                 SQL experiments на users/tasks
```

## Миграции и полная структура схемы

`migrations/001_init.sql` содержит `-- +goose Up`, `-- +goose Down` и пары
`StatementBegin` / `StatementEnd`, как в принятом SQL-формате. Транзакцией и
историей применений управляет goose; вручную BEGIN/COMMIT здесь не добавляем.

```bash
make up          # применить ещё не выполненные Up
make status      # applied/pending и текущая версия
make down        # откатить последнюю миграцию: tasks → users
make structure   # обновить structure.sql через pg_dump --schema-only
```

Команды используют goose v3.22.1 через `cmd/migrate`; отдельный CLI устанавливать
не нужно, Go остаётся 1.22. Старые `make db-up` / `make db-down` — алиасы up/down.
Не запускайте общий SQL-файл через psql -f: psql не различает секции goose и
выполнит обе. Используйте up/down.

В корневом [structure.sql](structure.sql) сразу видны таблицы users/tasks,
identity sequences, поля/типы, defaults, PK/UNIQUE/FK/CHECK, индекс, комментарии
и структура служебной goose_db_version. Данных пользователей и истории применений
в schema-only dump нет. Это справочный снимок, источником изменений остаются migrations.
Для обновления нужны pg_dump и DATABASE_URL; владельцы/ACL в снимок не включаются.
Сценарий перехода с уже применённого старого SQL без потери данных — в
[setup](docs/setup.md#если-старая-схема-уже-создана-через-psql).

## Быстрые и database tests

```bash
gofmt -w $(find . -name '*.go' -not -path './bin/*')
go test ./...
go vet ./...
go test -race ./...
make build
```

Эти проверки не подключаются к PostgreSQL. Существующие unit/HTTP tests
используют fake/memory, как в 13. Отдельно, с TEST_DATABASE_URL:

```bash
make test-db
make test-db-race
```

Тесты с `-tags=integration` создают свою схему, применяют ту же migration и
убирают только эту схему. Они не TRUNCATE-ят пользовательские таблицы.
Если DSN не задан, explicit integration run завершается ошибкой, а не Skip.
`make ci` — быстрые проверки; `make ci-db` добавляет database checks.

## SQL fundamentals на users/tasks

```bash
make sql-demo
# Дополнительно:
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -f sql/08_trigger.sql
```

Лаборатории показывают database/schema/table/row/column, типы/NULL, constraints,
INSERT/SELECT/UPDATE/DELETE/WHERE/JOIN, индекс/EXPLAIN, BEGIN/SAVEPOINT/ROLLBACK.
Query/QueryRow/Scan/Exec, параметры $1 и ErrNoRows разбираем в Repository.
Lab data и optional trigger откатываются. Выданные identity numbers могут иметь
пропуски: sequence allocation не откатывается вместе с INSERT, это нормально.

Порядок занятия — [live coding](docs/live-coding-script.md).
Детали — [SQL cheat sheet](docs/sql-cheatsheet.md),
[architecture](docs/architecture.md), [interview](docs/interview.md).
Homework и презентации не меняются; новые users/tasks HTTP features не добавляются.

## Логи

[Уровни DEBUG / INFO / ERROR / FATAL, LOG_LEVEL и безопасные поля](docs/logging.md).

`LOG_LEVEL=DEBUG make run` включает диагностику; обычный запуск использует INFO.
