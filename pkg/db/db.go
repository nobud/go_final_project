package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

const schema = `
	CREATE TABLE IF NOT EXISTS scheduler (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    date CHAR(8) NOT NULL DEFAULT "",
	    title VARCHAR(255) NOT NULL DEFAULT "",
	    comment TEXT,
	    repeat VARCHAR(128) NOT NULL DEFAULT ""
	);
	CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

func Init(dbFile string) error {
	dir := filepath.Dir(dbFile)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("ошибка создания директории %s: %w", dir, err)
		}
	}

	connection, err := sql.Open("sqlite", dbFile+"?cache=shared")
	if err != nil {
		return fmt.Errorf("ошибка открытия БД %w", err)
	}

	if _, err := connection.Exec(schema); err != nil {
		_ = connection.Close()
		return fmt.Errorf("ошибка создания схемы БД %w", err)
	}
	db = connection

	return nil
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
