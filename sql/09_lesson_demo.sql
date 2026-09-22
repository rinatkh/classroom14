\echo '1. Структура таблицы'
\d tasks

\echo '2. Создаём задачу'
INSERT INTO tasks (title)
VALUES ('Полный SQL-сценарий')
RETURNING id AS demo_task_id \gset

\echo 'Создана задача с id' :demo_task_id
SELECT id, title, done, created_at, updated_at
FROM tasks
WHERE id = :demo_task_id;

\echo '3. Читаем список'
SELECT id, title, done
FROM tasks
ORDER BY id DESC;

\echo '4. Меняем именно созданную задачу'
UPDATE tasks
SET done = true
WHERE id = :demo_task_id
RETURNING id, title, done, updated_at;

\echo '5. Проверяем план фильтра'
EXPLAIN
SELECT id, title
FROM tasks
WHERE done = true
ORDER BY created_at DESC;

\echo '6. Удаляем демонстрационную задачу'
DELETE FROM tasks
WHERE id = :demo_task_id
RETURNING id, title;
