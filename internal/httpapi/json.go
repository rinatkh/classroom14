package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

// decodeJSON читает ровно один JSON-объект.
//
// DisallowUnknownFields помогает заметить опечатку вроде titel вместо title.
// Вторая Decode должна получить io.EOF; иначе клиент прислал несколько JSON-
// значений в одном body, хотя договор endpoint ожидает одно.
func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("прочитать JSON: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body должен содержать один JSON-объект")
	}

	return nil
}

// writeJSON соблюдает порядок HTTP-ответа: header, status, body.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		// Status и часть body уже могли уйти клиенту. Здесь остаётся только логика
		// уровня логирования, которую добавим в следующих backend-занятиях.
		return
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
