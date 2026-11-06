# Что сломало бы continuity занятия14

## Снова Handler → Repository

Удаление UseCase убирает архитектурный результат 12–13. Берём финал13 целиком,
меняем только main и добавляем SQL implementations. Handler по-прежнему
зависит от TaskUseCase/AuthUseCase interfaces.

## pgx в application

Возврат pgx.Row из task.Repository меняет интерфейс под удобство драйвера и
делает fake тяжёлым. Repository возвращает прежние Task/User/error; Scan и
pgx.ErrNoRows живут в postgres package.

## Потеря auth ради первого SQL примера

Удалённые JWT/middleware/OwnerID или public Task routes — регрессия. Тот же
Alice/Bob curl flow должен работать с БД, включая 401 и 403.

## Authorization спрятана в SQL

Если Update просто добавил WHERE owner_id=current_user_id и вернул 404,
теряется объяснение 13. UseCase загружает Task и явно возвращает ErrForbidden;
SQL Update принимает прежний taskID и данные, ownership уже проверен.

## База получает HTTP context.Background

Так SQL продолжает ждать после отмены request. Передаём исходный ctx во все
Query/QueryRow/Exec; integration test отменяет ctx и проверяет context.Canceled.

## Склейка SQL с title/email

fmt.Sprintf("... '%s' ...", title) превращает данные в SQL program. Параметры
$1/$2 и отдельные args сохраняют кавычки/SQL-looking строки как данные.

## Unique precheck без constraint

Два Register могут одновременно не найти email. UNIQUE — финальная защита;
Repository переводит 23505 users_email_key в user.ErrAlreadyExists. SQL details
не попадают в клиентский error message.

## Все database errors становятся 404

ErrNoRows и FK/internal failure — разные причины. Только отсутствие записи
сопоставляем с ErrNotFound. Остальное остаётся внутренним 500 на HTTP boundary.

## Изменили порядок списка ради индекса

В 13 был id ASC. Индекс (owner_id,id) построен под этот реальный запрос.
Переход на created_at DESC был бы изменением поведения, не заменой persistence.

## Integration test очищает чужую БД

Не TRUNCATE-им общие users/tasks и не выполняем db-reset перед обычными unit
тестами. Tagged tests создают собственную схему из настоящих migrations.
Обычные go test/vet/race остаются быстрыми и работают без PostgreSQL.
