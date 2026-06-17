package main

import (
	"log"
	"os"

	"go_final_project/pkg/db"
	"go_final_project/pkg/server"
)

func main() {
	// путь к файлу БД можно переопределить через переменную окружения
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	// инициализируем базу данных (создаём таблицу, если нужно)
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("не удалось инициализировать БД: %v", err)
	}

	// запускаем веб-сервер
	server.Run()
}
