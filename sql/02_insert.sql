-- INSERT добавляет строки. RETURNING сразу показывает созданные данные,
-- включая ID и время, которые сформировал PostgreSQL.
INSERT INTO tasks (title)
VALUES
    ('Подключить PostgreSQL'),
    ('Написать SELECT'),
    ('Проверить Task API')
RETURNING id, title, done, created_at;

-- Укажем task_id через подзапрос, чтобы пример не зависел от конкретного ID.
INSERT INTO task_comments (task_id, body)
SELECT id, 'Первый комментарий к задаче'
FROM tasks
WHERE title = 'Подключить PostgreSQL'
ORDER BY id DESC
LIMIT 1
RETURNING id, task_id, body;
