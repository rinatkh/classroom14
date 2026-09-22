# Сценарий live coding

## До урока

1. Проверить `psql --version`.
2. Создать `go_course_tasks` и `go_course_tasks_test`.
3. Скопировать `.env.example` в `.env` и при необходимости изменить пароль.
4. Убедиться, что порт `8080` свободен.
5. Не применять миграцию заранее: первый `make db-up` выполняется вместе.

## Часть 1. Смотрим на базу

Откройте `migrations/001_init.up.sql`.

Спросите: «Какие поля из Go-структуры Task узнаваемы в таблице?» Ожидаемый
ответ: `id`, `title`, `done`; временные поля добавляет новая версия модели.

Выполните:

```bash
make db-up
psql "$DATABASE_URL"
```

В `psql`:

```text
\d tasks
\d task_comments
```

Покажите, что `PRIMARY KEY` и индексы — часть схемы, а не Go-кода.

## Часть 2. CRUD руками

Выполняйте `sql/02_insert.sql` — `sql/05_join.sql` частями. Перед `UPDATE` и
`DELETE` закройте `WHERE` рукой и спросите: «Сколько строк затронет команда без
этого условия?» Ожидаемый ответ: все строки.

Для `NULL` временно выполните:

```sql
SELECT NULL = NULL, NULL IS NULL;
```

Первый результат — `NULL`, второй — `true`.

## Часть 3. Индекс и план

Выполните `sql/06_index_explain.sql`. Если увидите `Seq Scan`, это полезный
результат: на маленькой таблице прочитать всё дешевле. Не обещайте, что наличие
индекса заставляет PostgreSQL его использовать.

## Часть 4. Транзакция и trigger

Выполните `BEGIN`, `UPDATE`, `SELECT`, `ROLLBACK`. Повторный `SELECT` покажет,
что изменение исчезло. Затем запустите `sql/08_trigger.sql` и сравните
`updated_at` до и после.

## Часть 5. Подключаем Go

Откройте по порядку:

1. `cmd/api/main.go` — строка подключения, pool, `Ping`, `Close`;
2. `internal/task/repository.go` — договор HTTP-слоя с хранилищем;
3. `internal/task/postgres.go` — SQL и `Scan`;
4. `internal/httpapi/handler.go` — тот же HTTP-контракт, что раньше.

Подчеркните: handler не знает SQL, а repository не знает status codes.

## Часть 6. Полный путь

Запустите `make run`, затем `make curl-demo` в соседнем терминале. После POST
откройте `psql` и выполните:

```sql
SELECT id, title, done FROM tasks ORDER BY id;
```

Перезапустите Go-сервер и снова сделайте GET. Это главный момент урока: процесс
перезапустился, а данные остались в отдельной системе — PostgreSQL.

## Часть 7. Тесты

```bash
make test
make test-db
```

В `handler_test.go` используется fake repository: тест проверяет HTTP отдельно.
В `postgres_integration_test.go` используется настоящая БД: тест проверяет SQL,
схему и преобразование строк в Go-структуры вместе. `make test-db` сначала
выполняет `test-db-reset`, поэтому миграция гарантированно применена именно к
`TEST_DATABASE_URL`.
