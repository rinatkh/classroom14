# Установка и первый вход в PostgreSQL

Для урока нужны сервер PostgreSQL 18 или 17 и консольная программа `psql`.
Docker не обязателен: цель занятия — сначала увидеть саму базу и подключение.

## macOS

Один из простых вариантов — Homebrew:

```bash
brew install postgresql@18
brew services start postgresql@18
psql postgres
```

## Ubuntu / Debian

Версия в системном репозитории может отличаться, для урока это допустимо:

```bash
sudo apt update
sudo apt install postgresql postgresql-client
sudo systemctl enable --now postgresql
sudo -u postgres psql
```

## Windows

Установите PostgreSQL с официального сайта, запомните пароль пользователя
`postgres` и добавьте каталог `bin` PostgreSQL в `PATH`. Затем откройте
PowerShell:

```powershell
psql -U postgres -d postgres
```

## Проверка

Внутри `psql`:

```sql
SELECT version();
\l
\q
```

Полезно различать:

- `postgres` — запущенный сервер базы данных;
- `psql` — клиент, который отправляет серверу SQL;
- `go_course_tasks` — конкретная база внутри сервера;
- `public` — схема, в которой на уроке создаются таблицы.

## Создание учебных баз

```bash
createdb -U postgres go_course_tasks
createdb -U postgres go_course_tasks_test
```

Если команда `createdb` недоступна, сделайте то же через `psql`:

```sql
CREATE DATABASE go_course_tasks;
CREATE DATABASE go_course_tasks_test;
```

Не используйте боевую или личную базу для `make db-reset`: эта команда удаляет
учебные таблицы перед повторным созданием.
