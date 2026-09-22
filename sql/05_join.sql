-- JOIN соединяет строку комментария со строкой его задачи по ключам.
SELECT
    t.id AS task_id,
    t.title,
    c.id AS comment_id,
    c.body
FROM tasks AS t
JOIN task_comments AS c ON c.task_id = t.id
ORDER BY c.id;

-- LEFT JOIN сохраняет даже задачи без комментариев.
SELECT
    t.id,
    t.title,
    count(c.id) AS comments_count
FROM tasks AS t
LEFT JOIN task_comments AS c ON c.task_id = t.id
GROUP BY t.id, t.title
ORDER BY t.id;
