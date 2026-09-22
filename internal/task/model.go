package task

import "time"

// Task — задача, которую API получает из PostgreSQL и возвращает клиенту.
//
// В прошлом занятии ID выдавал счётчик внутри Store. Теперь ID создаёт сама
// база данных. Временные поля тоже приходят из PostgreSQL: created_at задаёт
// DEFAULT, а updated_at обновляет trigger.
type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Changes описывает частичное изменение Task.
// nil означает «клиент не прислал поле», а указатель на значение означает
// «поле присутствовало в JSON и его нужно записать».
type Changes struct {
	Title *string
	Done  *bool
}
