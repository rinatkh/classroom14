# Continuity: Lesson13 final → Lesson14 final

## Исходная точка

Исходник — финальный classroom13, который открываем из настоящего соседнего
проекта. В classroom14 один Go-модуль, без вложенных копий уроков. Для показа
memory-поведения из корня classroom14: `make -C "../../13 Занятие/classroom13" run`.
При другом расположении репозиториев укажите свой путь к classroom13.
Остановите его API перед запуском classroom14 на том же порту.

## Не изменилось в финале14

| Группа | Проверяемое сохранение |
|---|---|
| internal/task | Model с OwnerID, inputs, errors, Repository interface, UseCase и все business tests |
| internal/user | User, AuthUseCase, Repository/Hasher/Issuer contracts, email normalization и tests |
| internal/auth | bcrypt/JWT, claims, signature/expiration checks и tests |
| internal/httpapi | routes, DTO, Mapper, handlers, errors mapping, middleware, authctx и tests |
| internal/repository/memory | прежние reference/fast-test implementations |
| test/integration | прежний memory-backed HTTP contract/lifecycle |
| scripts/curl-demo.sh | тот же полный auth/Task scenario, без правок |

Эти файлы совпадают с 13 после нормализации module/import prefix classroom13 →
classroom14. Не добавлено ни одного SQL/pgx import в UseCase/Handler и ни одного
net/http import в production task/user. Публичный JSON, PATCH false/nil,
id ASC, 401/403/404/409 и ownership сохранились.

## Изменилось

- cmd/api/main.go: DATABASE_URL, startup context/pool/Ping и два postgres constructors;
- internal/repository/postgres/db.go: инфраструктурная поверхность Pool/Tx;
- postgres/user.go: прежний user.Repository, normalised email/unique mapping;
- postgres/task.go: прежний task.Repository, SQL CRUD/List/CountActive;
- postgres/*_integration_test.go: новые real PostgreSQL проверки;
- migrations/001_init.sql: goose Up/Down, users/tasks, FK/CHECK/defaults/index;
- cmd/migrate/main.go: отдельные up/down/status через goose Provider;
- structure.sql и scripts/dump-structure.sh: полный schema-only snapshot и команда обновления;
- go.mod/go.sum: module classroom14, Go 1.22.0, pgx 5.7.2 вместо toolchain upgrade;
- Makefile/config/CI/docs/sql labs: воспроизводимый запуск и отдельные DB checks.

**Какие унаследованные production/application файлы13 изменены ради БД?**
По логике только cmd/api/main.go. В остальных Go-файлах — только техническая
смена import prefix, необходимая для нумерации classroom14. Новые SQL
implementations размещены отдельно, interfaces под pgx не переделаны.

## Из прежнего classroom14 удалён откат архитектуры

Старые Handler → Repository и domain-level PostgresStore заменены сохранённым
финалом13. Нет потери UseCase/auth/ownership. task_comments удалена: PK/FK/JOIN
показаны на естественных users/tasks. Нет classroom15 или go 1.25 в актуальном
root module/CI. Production trigger не устанавливается; есть отдельный rollback lab.

## Runtime доказательство

1. classroom13: register/login/create, restart → login 401 и пустые Task после новой регистрации.
2. after: тот же flow с PostgreSQL, restart без make down → login/GET снова 200.
3. Bob после restart всё ещё получает 403 на чужие GET/PATCH/DELETE.
4. pgx.ErrNoRows и unique email дают прежние 404/409; internal ошибки скрыты.
5. Быстрые tests работают без БД; tagged tests проверяют SQL/constraints/rollback/migrations.

Homework и презентации не изменяются. Тема14 — persistence существующего
приложения, а не ещё один вариант похожего Task API.

## Наследование логов

Единый `internal/logging` появился в classroom10. DEBUG / INFO / ERROR / FATAL,
`LOG_LEVEL` и безопасные поля сохраняются в финальных версиях следующих
занятий без вложенных проектов. Меняется набор наблюдаемых операций, прежний Task HTTP contract остаётся.
Подробности и показ уровней — [logging.md](logging.md).
