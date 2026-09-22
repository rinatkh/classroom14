-- Все команды между BEGIN и COMMIT составляют одну операцию.
BEGIN;

UPDATE tasks
SET done = true
WHERE title = 'Подключить PostgreSQL';

SELECT id, title, done
FROM tasks
WHERE title = 'Подключить PostgreSQL';

-- На уроке сначала выполните ROLLBACK и убедитесь, что изменение исчезло.
ROLLBACK;

-- Затем повторите BEGIN + UPDATE и завершите COMMIT, чтобы сохранить изменение.
