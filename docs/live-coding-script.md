# Live coding 14 — реализация Repository меняется, приложение остаётся

Ориентир 90–110 минут. Core SQL и подключение — основной маршрут; EXPLAIN и
transaction basics можно показать готовыми короткими опытами. Trigger — optional.
Ничего из архитектуры/auth прошлого урока не выкидываем.

## STEP 0. Показываем проблему памяти

Запустите classroom13/cmd/api через `make -C "../../13 Занятие/classroom13" run` с JWT_SECRET. Выполните
register Alice, login, POST /tasks. Сохраните полученный ID и token. Команды
в [persistence-demo.md](persistence-demo.md).

Остановите API classroom13, начните новый memory-прогон со свежим ключом, как в 13.
Login Alice теперь 401: User потерян. Повторный register успешен; после login
GET /tasks возвращает [], GET прежней Task — 404. Речь о данных, не о JWT bug.
Остановите API classroom13, чтобы освободить порт.

## STEP 1. Существующие interfaces

Откройте classroom13/internal/user/repository.go и classroom13/internal/task/repository.go.
Какие слои должны измениться при правильной архитектуре? Ответ: concrete
implementations и wiring. Handler/UseCase/DTO/Mapper/auth/ownership не трогаем.
Сопоставьте эти же interfaces в корне: сигнатуры прежние.

## STEP 2. PostgreSQL и migrations

По setup создайте server/database, задайте DATABASE_URL. Разведите database,
schema, table, row, column. Примените `make up`, не пишите таблицы отдельно
вручную. Up — воспроизводимая версия схемы, down её откатывает.

## STEP 3. Users

В migrations/001_init.sql разберите id bigint identity PK, email UNIQUE NOT
NULL, password_hash NOT NULL, created_at timestamptz DEFAULT now(). Email уже
normalised AuthUseCase. Extension не нужен. PasswordHash остаётся непрозрачной
строкой от неизменённого bcrypt, не plaintext.

## STEP 4. Tasks и owner FK

Разберите users 1:N tasks, owner_id REFERENCES users(id), NOT NULL, CHECK title,
done DEFAULT false и timestamps. ON DELETE CASCADE выбран для удаления учебных
Task вместе с User; обсудите сохранение истории как иной продуктовый выбор.
Task.OwnerID уже существовал в 13 — новый домен не придумываем.

## STEP 5. Ручные INSERT/SELECT

```bash
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -f sql/01_inspect_schema.sql
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -f sql/02_insert.sql
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -f sql/03_select.sql
```

Для core показа достаточно sql/01_inspect_schema.sql, 02_insert.sql,
03_select.sql, 04_update_delete.sql и 05_join.sql. Установите рядом с SQL User
и Task из Go. Покажите типы/NULL, WHERE, RETURNING и JOIN users/tasks.
Данные lab откатываются; выдача identity numbers может оставить пропуски.
SQL выполняет администратор БД, а права Alice/Bob на HTTP по-прежнему проверяет UseCase.

## STEP 6. PostgresUserRepository

Откройте internal/repository/postgres/user.go. Реализуйте прежние Create и
GetByEmail. QueryRow возвращает одну строку, Scan получает результат/ошибку.
Покажите $1/$2 и аргументы отдельно от SQL. pgx.ErrNoRows превращается в
user.ErrNotFound; 23505 users_email_key — в user.ErrAlreadyExists. Constraint
защищает одновременные Register после одинакового precheck.

## STEP 7. PostgresTaskRepository

В task.go реализуйте все прежние Create/Get/List/Update/Delete/CountActive.

- Query → Rows/Next/Scan/Close/Err для списка;
- QueryRow → Scan для одной записи и COUNT;
- Exec → RowsAffected для удаления;
- COALESCE сохраняет partial PATCH и явный false;
- Get не фильтрует current user: ownership остаётся в UseCase.

Найдите каждый ctx: передаём тот, что пришёл от Handler. Нет Background внутри SQL.
Индекс (owner_id,id) обслуживает прежний owner filter + id ASC; не меняйте sort.

## STEP 8. Меняем main wiring

Откройте diff main из README. Создайте pool, выполните Ping с startup timeout,
передайте pool двум postgres constructors. Всё ниже — прежние hasher, tokens,
UseCases, handlers, middleware, router. Найдите, что в UseCase не добавился pgx.

## STEP 9. Тот же HTTP lifecycle

С DATABASE_URL и JWT_SECRET выполните `make run`, затем `make curl-demo` во
втором терминале. Сохраняются register/login, protected Task CRUD, прежние DTO
и ошибки. Дополнительно создайте отдельную Task для restart, не удаляя её.

## STEP 10. Restart

Остановите только API. С тем же ключом и тем же DATABASE_URL снова `make run`.
Login Alice снова 200, собственные Task остались. Открытый ранее token до exp
тоже работает: JWT не переписывали, User ID больше не теряется.

## STEP 11. Bob и 403

Зарегистрируйте Bob и получите его token. Bob GET/PATCH/DELETE Task Alice → 403;
Bob GET /tasks не содержит Alice Task. Откройте прежний TaskUseCase.Get и
сравнение OwnerID: решение принимается там же, SQL только загрузил запись.

## STEP 12. Tests и завершение

```bash
make ci
make test-db
make test-db-race
```

Старые tests по-прежнему используют fake/memory. Отдельные tagged tests
поднимают schema из migrations и проверяют real repositories, FK/CASCADE,
unique, cancellation, rollback/up-down-up и полный auth → ownership путь.

После core покажите sql/04_update_delete.sql, sql/05_join.sql,
sql/06_index_explain.sql и sql/07_transaction.sql. `make sql-demo` выполняет
весь этот набор для повторения.
Дополнительный sql/08_trigger.sql демонстрирует database-side behavior внутри
ROLLBACK; production updated_at обновляется обычным UPDATE Repository.

Финальный вопрос: «Какие файлы application пришлось переписать ради БД?»
Ответ: только main по логике; новые persistence implementations, migrations и
config/deps. Handler, UseCase, JWT, middleware и authorization не переписаны.
