# Setup занятия 14

Нужны Go 1.22.x, PostgreSQL 16+, psql, pg_dump и Python 3 для учебных scripts.
Пользуйтесь уже установленной подходящей
версией PostgreSQL; установка нового Go toolchain не является частью урока.
Официальные установщики и инструкции: [PostgreSQL downloads](https://www.postgresql.org/download/).
На Windows добавьте bin PostgreSQL в PATH; make/curl scripts выполняются в
привычном для курса Bash/WSL. Docker для локального занятия не обязателен.

## Роль PostgreSQL и две учебные базы

Подключитесь администратором **локальной учебной** установки:

```bash
psql -h localhost -U postgres -d postgres
```

Один раз создайте роль и базы:

```sql
CREATE ROLE course_user LOGIN PASSWORD 'course-local-password';
CREATE DATABASE go_course_tasks OWNER course_user;
CREATE DATABASE go_course_tasks_test OWNER course_user;
```

course_user — техническая роль подключения сервиса к БД. Alice/Bob из users —
business users нашего API, а не дополнительные PostgreSQL login roles.
Учебный пароль выше используется только в локальном примере. При других
параметрах установки задайте свои значения в DSN.

## Окружение

```bash
export DATABASE_URL='postgres://course_user:course-local-password@localhost:5432/go_course_tasks?sslmode=disable'
export TEST_DATABASE_URL='postgres://course_user:course-local-password@localhost:5432/go_course_tasks_test?sslmode=disable'
export JWT_SECRET="$(openssl rand -hex 32)"
```

HOST/PORT могут отличаться. sslmode=disable здесь относится только к локальному
учебному соединению. Makefile не подхватывает .env: он получает export-нутые
значения. .env.example — справочник. Секрет не выводите в console log.

Проверка подключения:

```bash
psql "$DATABASE_URL" -X -c 'SELECT current_database(), current_schema(), version();'
```

## Воспроизводимая схема

```bash
make up
make status
make run
```

Goose выполняет секцию Up из migrations/001_init.sql: users → tasks → индекс.
В том же файле находятся Down и StatementBegin/StatementEnd. История хранится
в goose_db_version. Повторный make up не исполняет уже применённый DDL и
не очищает данные. Таблицы вручную не создаём. Отдельный goose CLI не нужен:
cmd/migrate использует зафиксированную Go-зависимость v3.22.1, совместимую с Go 1.22.
Перезапуск API использует ту же БД и тот же JWT_SECRET.

Команда `make down` явно удаляет учебные users/tasks и их данные, затем
`make up` снова создаёт пустую схему. Это демонстрация migration rollback,
а не обычная операция restart приложения.
После полного сброса users/tasks начинайте новый учебный прогон со свежим
JWT_SECRET: временные ID могут выдаваться заново. При обычном restart без
сброса данных ключ сохраняем.

Можно вызвать команды напрямую: `go run ./cmd/migrate up`, `down` или `status`.
Весь migrations/001_init.sql через psql не запускаем: psql выполнит и Up, и Down.
Транзакцию для каждой секции открывает goose, поэтому собственных BEGIN/COMMIT
в миграции нет. Down откатывает одну последнюю применённую версию.

## Полная схема в structure.sql

В корне уже есть schema-only snapshot из pg_dump: таблицы, identity sequences,
типы/defaults, ограничения, индекс и комментарии. Включена и структура
goose_db_version; строки пользователей, задач и истории миграций не выгружаются.

```bash
make structure
# При нескольких версиях клиента можно указать нужный pg_dump:
make structure PG_DUMP=/path/to/postgresql/bin/pg_dump
```

pg_dump должен поддерживать версию подключённого PostgreSQL. Используются
--schema-only --no-owner --no-privileges --schema=public. При ошибке существующий
structure.sql не затирается. Снимок нужен для просмотра/сравнения структуры;
для развёртывания базы используем make up, а не импорт snapshot вместо миграций.
После следующей миграции обновите snapshot командой make structure.

## Если старая схема уже создана через psql

Если ранее был выполнен именно прежний migrations/001_init.up.sql этого
classroom, users/tasks уже существуют, но истории goose ещё нет. Состав DDL
в новой секции Up тот же. Не выполняйте down ради перехода: это удалит данные.

После проверки, что схема действительно соответствует прежней 001, создайте
историю и отметьте эту уже выполненную версию, не повторяя DDL:

```bash
make status
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c \
  "INSERT INTO goose_db_version(version_id,is_applied) SELECT 1,true WHERE NOT EXISTS (SELECT 1 FROM goose_db_version WHERE version_id=1 AND is_applied);"
make status
make up
```

Этот ручной baseline подходит только для той же старой 001, не для произвольной
чужой схемы. На новой пустой базе baseline не нужен: сразу make up. Новые
команды не пытаются автоматически пометить найденные таблицы как мигрированные.

## Тестовая БД

```bash
make test-db
make test-db-race
```

Integration fixture запускает настоящий goose Provider и применяет 001 в отдельной случайной схеме внутри
TEST_DATABASE_URL; роли нужны права создания/удаления собственных схем.
Таблицы основной схемы не очищаются. Обычные `make ci` и `go test ./...`
работают без DSN и без запущенного PostgreSQL.

Формат секций: [goose annotations](https://pressly.github.io/goose/documentation/annotations/).
Используемый API: [goose Provider](https://pressly.github.io/goose/documentation/provider/).
