-- EXPLAIN показывает план, но не выполняет SELECT.
-- На маленькой таблице PostgreSQL может честно выбрать Seq Scan: прочитать
-- несколько строк целиком дешевле, чем обращаться к индексу.
EXPLAIN
SELECT id, title, created_at
FROM tasks
WHERE done = false
ORDER BY created_at DESC;

-- EXPLAIN ANALYZE действительно выполняет запрос и измеряет его.
-- Для UPDATE/DELETE используйте его только внутри откатываемой транзакции.
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, title, created_at
FROM tasks
WHERE done = false
ORDER BY created_at DESC;
