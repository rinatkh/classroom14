# Security сохранилась при смене storage

JWT, bcrypt, auth middleware, typed request context и ownership код совпадают
с финалом13 после module/import rename. SQL Repository не принимает решение,
имеет ли Bob право изменить Task Alice. Его принимает прежний TaskUseCase.

users.password_hash содержит результат прежнего PasswordHasher; password/hash
не публикуются DTO или JWT claims. PostgreSQL User Repository возвращает
user.ErrAlreadyExists для email unique violation без текста PgError. Внутренний
FK/SQL failure получает существующий нейтральный HTTP 500.

DATABASE_URL и JWT_SECRET получаются из окружения, в startup log не печатаются.
В отличие от memory demo, обычный restart финала14 сохраняет и данные, и user ID:
используйте прежний ключ, чтобы действующие tokens сохраняли смысл. Если ключ
изменить, старые подписи не пройдут — это поведение JWT, не потеря persistence.
При полном сбросе учебных таблиц через make down / make up начинайте новый прогон
со свежим ключом, поскольку identity numbers в новой схеме могут выдаваться заново.

Этот урок не добавляет refresh/session/revocation. До exp access JWT ведёт себя
так же, как в 13. Локальные DSN с sslmode=disable относятся к учебной установке.
