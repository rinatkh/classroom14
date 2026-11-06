# Вопросы первого PostgreSQL-урока

1. **Какие слои изменились?** Только concrete persistence и composition root;
   прежние UseCase/HTTP/auth interfaces не переписаны.
2. **Зачем вводили Repository в 12?** Чтобы именно эту смену реализации делать
   без переноса business logic в infrastructure.
3. **Что теперь переживает restart?** Users, Tasks и выданные им ID в БД.
4. **Почему ключ JWT оставляем тем же между restart?** Persistence не меняет
   подпись/TTL, identity теперь ссылается на сохранённые user ID.
5. **Database/schema/table — одно?** Database содержит namespaces схем;
   schema содержит таблицы, таблица — строки с колонками.
6. **PK отличается от UNIQUE?** PK задаёт идентичность и NOT NULL; UNIQUE
   выражает дополнительную уникальность, например email.
7. **Что гарантирует FK?** Owner задачи существует. Это не HTTP authorization.
8. **Почему CASCADE?** Выбрали политику удалить учебные Task вместе с User;
   требование хранить историю привело бы к другому решению.
9. **Зачем CHECK при проверке UseCase?** Он защищает данные от записи в обход
   приложения. Бизнес-решения приложения всё равно остаются в UseCase.
10. **Почему email без extension?** AuthUseCase уже нормализует, Repository
    сохраняет/ищет normalised key, UNIQUE защищает хранимые значения.
11. **Что делает default?** Подставляет done=false/now(), если значение не передано.
12. **NULL равен NULL?** Обычное сравнение даёт unknown; используем IS NULL.
13. **Когда QueryRow/Query/Exec?** Одна строка, несколько строк, команда без
    требуемых строк ответа соответственно; проверяем Scan/Rows.Err/RowsAffected.
14. **Почему QueryRow error проверяется на Scan?** Получение строки/ошибки
    завершается при чтении результата.
15. **Как сохранить 404?** Преобразовать pgx.ErrNoRows в прежний ErrNotFound;
    для DELETE проверить RowsAffected.
16. **Где 403?** Там же в TaskUseCase: OwnerID сравнивается с userID.
17. **Как не потерять partial PATCH?** Оставить pointer inputs; nil через
    COALESCE не меняет поле, &false записывает false.
18. **Что даёт параметризация?** SQL и данные разделены, кавычки в title не
    становятся частью исполняемой команды.
19. **Почему индекс owner_id/id?** Он соответствует реальному фильтру и
    сохранённому order. Индекс на bool без query pattern не нужен.
20. **Почему Seq Scan бывает нормальным?** На малом наборе это может быть
    дешевле; EXPLAIN нужно читать вместе со статистикой и размерами данных.
21. **Зачем JOIN?** Получить связанные Task и email User; новый HTTP endpoint
    ради JOIN в этом уроке не нужен.
22. **Что меняет transaction?** Группирует операции в all-or-nothing;
    rollback откатывает данные, но не обещает непрерывных identity IDs.
23. **Достаточен ли process mutex для нескольких API instances?** Нет.
    В этом уроке сохраняем прежний single-process сценарий; distributed
    atomicity не появляется от самого подключения pool.
24. **Почему быстрые tests не требуют БД?** Их dependencies прежние fake/memory;
    реальные SQL queries проверяются отдельным explicit integration target.
25. **Как доказать, что архитектура не откатилась?** Сравнить файлы с 13 после
    нормализации imports и повторить тот же HTTP/auth/ownership lifecycle.
