SELECT id, title, updated_at
FROM tasks
ORDER BY id
LIMIT 1;

-- Trigger сработает автоматически перед UPDATE и заменит updated_at.
UPDATE tasks
SET title = title || ' — изменено'
WHERE id = (SELECT min(id) FROM tasks)
RETURNING id, title, updated_at;
