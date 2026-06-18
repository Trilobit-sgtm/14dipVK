package server

import (
	"log"
	"net/http"
	"os"

	"go_final_project/pkg/api"
)

const defaultPort = "7540"
const webDir = "./web"

// Run регистрирует обработчики и запускает HTTP-сервер.
func Run() {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}

	// регистрируем все API-обработчики
	api.Init()

	// файл-сервер для фронтенда: отдаёт index.html, js, css, favicon
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	log.Printf("сервер запущен на порту %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}
}
