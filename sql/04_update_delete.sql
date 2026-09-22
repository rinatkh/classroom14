-- UPDATE без WHERE изменил бы ВСЕ строки. Перед выполнением проговорите условие.
UPDATE tasks
SET done = true
WHERE title = 'Написать SELECT'
RETURNING id, title, done, updated_at;

-- DELETE тоже требует осознанного WHERE. RETURNING показывает удалённую строку.
DELETE FROM tasks
WHERE title = 'Проверить Task API'
RETURNING id, title;
