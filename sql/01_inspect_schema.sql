-- \d — команда psql, а не SQL. Она показывает колонки, типы и ограничения.
\d tasks
\d task_comments

-- Комментарии к таблицам и колонкам PostgreSQL хранит в системном каталоге.
SELECT
    obj_description('tasks'::regclass) AS table_comment,
    col_description('tasks'::regclass, 2) AS title_comment;
