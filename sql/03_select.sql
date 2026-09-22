-- SELECT читает только перечисленные колонки. Явный список надёжнее SELECT *:
-- результат не изменится случайно после добавления новой колонки в таблицу.
SELECT id, title, done, created_at, updated_at
FROM tasks
ORDER BY id DESC;

-- WHERE оставляет только строки, которые подходят под условие.
SELECT id, title
FROM tasks
WHERE done = false
ORDER BY created_at DESC
LIMIT 2;
