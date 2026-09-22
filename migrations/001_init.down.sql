-- При удалении tasks PostgreSQL удалит принадлежащий ей trigger.
-- Таблицы удаляем раньше функции, от которой trigger зависит.
DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS tasks;
DROP FUNCTION IF EXISTS set_tasks_updated_at();
