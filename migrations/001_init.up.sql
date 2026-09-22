-- Таблица tasks хранит состояние Task API после перезапуска Go-сервера.
CREATE TABLE tasks (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title      text NOT NULL CHECK (btrim(title) <> ''),
    done       boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE tasks IS 'Задачи учебного Task API';
COMMENT ON COLUMN tasks.id IS 'ID, который выдаёт PostgreSQL';
COMMENT ON COLUMN tasks.title IS 'Название задачи, непустое после удаления пробелов';
COMMENT ON COLUMN tasks.done IS 'Признак выполнения задачи';
COMMENT ON COLUMN tasks.created_at IS 'Время создания записи';
COMMENT ON COLUMN tasks.updated_at IS 'Время последнего изменения записи';

-- Вторая таблица нужна, чтобы увидеть связь один-ко-многим и внешний ключ.
CREATE TABLE task_comments (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id    bigint NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    body       text NOT NULL CHECK (btrim(body) <> ''),
    created_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE task_comments IS 'Комментарии к задачам';
COMMENT ON COLUMN task_comments.task_id IS 'Ссылка на tasks.id';

-- Индекс помогает запросу быстро находить комментарии одной задачи.
-- PostgreSQL автоматически создаёт индекс для PRIMARY KEY, но не создаёт
-- отдельный индекс на колонку внешнего ключа task_id.
CREATE INDEX task_comments_task_id_idx ON task_comments (task_id);

-- Этот составной индекс соответствует учебному запросу:
-- WHERE done = $1 ORDER BY created_at DESC.
CREATE INDEX tasks_done_created_at_idx ON tasks (done, created_at DESC);

-- Trigger-функция записывает новое время перед каждым UPDATE строки tasks.
CREATE FUNCTION set_tasks_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER tasks_set_updated_at
BEFORE UPDATE ON tasks
FOR EACH ROW
EXECUTE FUNCTION set_tasks_updated_at();

COMMENT ON FUNCTION set_tasks_updated_at() IS
    'Обновляет tasks.updated_at перед изменением строки';
