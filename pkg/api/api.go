package api

import "net/http"

// DateLayout — формат даты, используемый во всём API
const DateLayout = "20060102"

// Init регистрирует все обработчики API-запросов.
// Вызывается из server.Run() до запуска сервера.
func Init() {
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/task/done", doneHandler)
}
