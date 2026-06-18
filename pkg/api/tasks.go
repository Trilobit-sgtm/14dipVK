package api

import (
	"net/http"

	"go_final_project/pkg/db"
)

// tasksResp — обёртка для JSON-ответа со списком задач
type tasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler обрабатывает GET /api/tasks.
// Возвращает список ближайших задач, отсортированных по дате.
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения списка задач: "+err.Error())
		return
	}
	writeJSON(w, tasksResp{Tasks: tasks})
}
