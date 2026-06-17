package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

// taskHandler — роутер для /api/task, разбирает запросы по HTTP-методу.
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeError(w, fmt.Sprintf("метод %s не поддерживается", r.Method))
	}
}

// ---- POST /api/task — добавление задачи ----

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, "ошибка разбора JSON: "+err.Error())
		return
	}

	if err := validateTask(&task); err != nil {
		writeError(w, err.Error())
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeError(w, "ошибка добавления задачи: "+err.Error())
		return
	}

	writeJSON(w, map[string]string{"id": fmt.Sprint(id)})
}

// ---- GET /api/task?id=<id> — получение задачи по id ----

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "не указан идентификатор задачи")
		return
	}

	task, err := db.GetTask(id)
	if err == sql.ErrNoRows {
		writeError(w, "задача не найдена")
		return
	}
	if err != nil {
		writeError(w, "ошибка получения задачи: "+err.Error())
		return
	}

	writeJSON(w, task)
}

// ---- PUT /api/task — обновление задачи ----

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, "ошибка разбора JSON: "+err.Error())
		return
	}

	if task.ID == "" {
		writeError(w, "не указан идентификатор задачи")
		return
	}

	if err := validateTask(&task); err != nil {
		writeError(w, err.Error())
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, map[string]string{})
}

// ---- DELETE /api/task?id=<id> — удаление задачи ----

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "не указан идентификатор задачи")
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, map[string]string{})
}

// ---- validateTask — общая проверка полей задачи ----

func validateTask(task *db.Task) error {
	if task.Title == "" {
		return fmt.Errorf("не указан заголовок задачи")
	}

	now := time.Now()

	// если дата не указана — берём сегодняшнее число
	if task.Date == "" {
		task.Date = now.Format(DateLayout)
	}

	// проверяем формат даты
	t, err := time.Parse(DateLayout, task.Date)
	if err != nil {
		return fmt.Errorf("дата указана в неверном формате, ожидается YYYYMMDD")
	}

	// если указано правило повторения — проверяем его корректность
	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("неверное правило повторения: %v", err)
		}
	}

	// если дата задачи меньше сегодняшней — скорректируем
	if dateOnly(t).Before(dateOnly(now)) {
		if task.Repeat == "" {
			// одноразовая задача в прошлом → ставим сегодня
			task.Date = now.Format(DateLayout)
		} else {
			// периодическая → берём уже вычисленную следующую дату
			task.Date = next
		}
	}

	return nil
}
