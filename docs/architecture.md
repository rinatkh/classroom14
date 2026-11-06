# Архитектурный payoff: сменили persistence implementation

Прежнее направление зависимостей сохраняется:
Handler → TaskUseCase interface → task.Repository.
AuthHandler → AuthUseCase interface → user.Repository / PasswordHasher / TokenIssuer.
AuthMiddleware → TokenVerifier. Mapper и DTO остаются на HTTP boundary.

Из новых concrete types main выбирает postgres.UserRepository и
postgres.TaskRepository. Оба реализуют **исходные interfaces занятия13**;
compile-time assertions подтверждают совпадение сигнатур.

## SQL boundary

| Файл/пакет | Изменение ради PostgreSQL |
|---|---|
| cmd/api/main.go | DB connection/Ping/pool, выбор двух repositories |
| internal/repository/postgres | новая SQL implementation |
| migrations, config, deps | схема и подключение |
| internal/task, internal/user | без логических изменений |
| httpapi Handler/DTO/Mapper/router | без логических изменений |
| auth, middleware, context | без логических изменений |

Единственный изменённый унаследованный production Go-файл по логике — main.
Остальные отличаются лишь module/import prefix classroom13 → classroom14.
Memory implementations и прежние tests сохранены, production wiring их не использует.

## Data vs authorization

List(ctx, ownerID) и CountActive(ctx, ownerID) фильтруют нужный набор данных.
Get(ctx, taskID) загружает запись, не знает current user. В прежнем UseCase
после Get сравниваются Task.OwnerID и userID; Update/Delete вызывают эту проверку.
SQL не подменяет ErrForbidden «записью не найдено», поэтому Bob по-прежнему
получает 403 для существующей чужой Task.

Update меняет только title/done, не OwnerID. nil pointer превращается в SQL NULL,
COALESCE сохраняет старое значение; &false остаётся false. trim/empty title,
no changes, repeated completion и лимит остались в UseCase.

## Context и ошибки

Handler передаёт r.Context() в UseCase, затем тот же ctx приходит в
QueryRow/Query/Exec. Background используется только при startup и в тестовом
fixture, не внутри SQL Repository.

pgx.ErrNoRows → task.ErrNotFound/user.ErrNotFound. Unique violation конкретного
users_email_key → точный user.ErrAlreadyExists без raw SQL details. Остальные
ошибки оборачиваются через %w; существующий Handler возвращает нейтральный 500.

DBTX в postgres/db.go — маленькая инфраструктурная поверхность Pool/pgx.Tx,
нужная SQL-коду и transaction lab. Она не меняет application Repository contract.

## Граница текущего урока

Один процесс и один UseCase сохраняют прежнюю защиту CountActive → запись
через changeMu. Смена map на SQL сама по себе не превращает это в межпроцессную
транзакцию для нескольких экземпляров приложения. На первом уроке изучаем
BEGIN/COMMIT/ROLLBACK отдельно; business layer ради демонстрации не переписываем.

HTTP health остаётся проверкой процесса, как в 13. Пользователи, Task и sequence
теперь долговечны; JWT/key и весь authentication path остались прежними.
