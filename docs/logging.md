# Логи Task API

Логгер появился в занятии 10 и остаётся тем же при развитии приложения.
Используем `log/slog` из стандартной библиотеки Go 1.22. Зависимостей для логов нет.
`internal/logging` настраивает stderr, `service`, уровни и формат один раз в main.

| Уровень | Когда | Пример |
| --- | --- | --- |
| DEBUG | Детали запроса, ожидаемый отказ 400/401/403/404, чтение данных | operation=task.get, task_id=3 |
| INFO | Запуск процесса и успешное изменение состояния | task created, task_id=3 |
| WARN | Стандартный уровень slog для восстанавливаемых предупреждений | Включён в фильтр, искусственно не генерируется |
| ERROR | Неожиданный сбой операции, после которого процесс ещё работает | Не удалось записать HTTP response или обработать событие |
| FATAL | Ошибка, после которой процесс не может продолжить работу | Неверная config, занятый обязательный listener |

По умолчанию `LOG_LEVEL=INFO`, `LOG_FORMAT=text`. DEBUG скрыт. Доступен JSON:

```bash
LOG_LEVEL=DEBUG LOG_FORMAT=text make run
LOG_LEVEL=INFO LOG_FORMAT=json make run
```

`LOG_LEVEL` принимает DEBUG, INFO, WARN, ERROR, FATAL без учёта регистра.
Фильтр оставляет выбранный уровень и более серьёзные. Неизвестное значение —
ошибка запуска с FATAL и exit code 1. WARN сохранён как стандартный уровень slog.
FATAL не встроен в slog: `logging.LevelFatal` задаёт имя, `logging.Fatal` пишет
record и вызывает `os.Exit(1)`. Он вызывается только в main после возврата run.
Ресурсы закрываются через defer внутри run до выхода. Обычная ошибка запроса
не завершает API. Fatal не применяется в handler, repository или business UseCase.

## Что пишем

Название операции, status, числовые task_id/user_id, количество объектов, время,
service. В сетевых лабораториях также размер сообщения и имя эксперимента.
В Kafka adapter — event type, event_id, task_id, partition/offset и стадия обработки.
Ожидаемый отказ клиента пишем в DEBUG, а не выдаём каждый 401/404 за серверный ERROR.
Один неожиданный сбой логируется на той границе, которая его обрабатывает.
Business UseCases возвращают ошибки и не зависят от конкретного logger.

## Что исключаем

Нельзя логировать body/title/email/password, Authorization/Bearer, cookies, JWT,
refresh token, token hash, Redis keys с session hash, DATABASE_URL/REDIS_URL с паролем.
Не выводим целиком request/response, input DTO или произвольный driver error.
`logging.Error` сохраняет тип ошибки без её raw text, в котором могут быть данные.
В startup сообщениях оставляем понятную безопасную причину и имя переменной,
но не её значение. Учебный stdout сетевых примеров — демонстрация протокола,
он отделён от служебных логов в stderr и не применяется к auth requests.

## Показ на занятии

1. Запустить API с INFO, вызвать health и создать Task: виден запуск и task created.
2. Перезапустить с DEBUG, повторить запросы: видны operation и ожидаемые отказы.
3. `go test -v ./internal/httpapi/... -run TestResponseFailureLogsWithoutLeakingData`
   проверяет настоящий write failure: ERROR записан, секреты не попали в log,
   уже отправленный HTTP status не переписывается.
4. `LOG_LEVEL=invalid go run ./cmd/api` демонстрирует FATAL до подключения к storage.
   Это отдельный процесс, рабочий сервер не требуется останавливать.
5. `go test -v ./internal/logging` проверяет фильтр, JSON, отсутствие секретов и
   FATAL/exit 1 в subprocess. Наличие DEBUG в исходнике не означает вывод при INFO.

Логи дополняют прежний HTTP contract, не меняют endpoints, JSON, PATCH, ownership
или storage. Middleware ради логов в первый HTTP-урок не добавляем.

Источники: https://pkg.go.dev/log/slog@go1.22.12 и
https://go.dev/src/log/slog/example_custom_levels_test.go.
