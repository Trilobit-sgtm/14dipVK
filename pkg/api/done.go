package api

import (
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

// doneHandler обрабатывает POST /api/task/done?id=<id>.
// Для периодической задачи пересчитывает дату; для одноразовой — удаляет.
func doneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "не указан идентификатор задачи")
		return
	}

	// получаем задачу из БД
	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}

	// одноразовая задача — просто удаляем
	if task.Repeat == "" {
		if err = db.DeleteTask(id); err != nil {
			writeError(w, http.StatusInternalServerError, "ошибка удаления задачи: "+err.Error())
			return
		}
		writeJSON(w, map[string]string{})
		return
	}

	// периодическая задача — вычисляем следующую дату и обновляем
	next, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ошибка вычисления следующей даты: "+err.Error())
		return
	}

	if err = db.UpdateDate(id, next); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка обновления даты задачи: "+err.Error())
		return
	}

	writeJSON(w, map[string]string{})
}
