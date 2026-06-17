package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// db — глобальное соединение с базой данных, используется во всём пакете
var db *sql.DB

// schema — идемпотентные SQL-команды для создания таблицы и индекса.
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    date    CHAR(8)      NOT NULL DEFAULT "",
    title   VARCHAR(256) NOT NULL DEFAULT "",
    comment TEXT         NOT NULL DEFAULT "",
    repeat  VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`

// Init открывает (или создаёт) файл БД и при необходимости разворачивает схему.
func Init(dbFile string) error {
	conn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if err = conn.Ping(); err != nil {
		return err
	}

	db = conn

	// Схема идемпотентная: это чинит случай, когда файл БД есть, а таблицы ещё нет.
	if _, err = db.Exec(schema); err != nil {
		return err
	}

	return nil
}
