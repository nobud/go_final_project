package db

import (
	"database/sql"
	"fmt"
	"os"

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
	CREATE INDEX idx_scheduler_date ON scheduler(date);
`

func Init(dbFile string) error {
	var install bool
	_, err := os.Stat(dbFile)

	if err != nil {
		install = true
	}

	db, err = sql.Open("sqlite", dbFile+"?cache=shared")
	if err != nil {
		return fmt.Errorf("ошибка открытия БД %w", err)
	}

	if install {
		_, err := db.Exec(schema)
		if err != nil {
			db.Close()
			return fmt.Errorf("ошибка создания схемы БД %w", err)
		}
	}

	db.SetMaxOpenConns(1)
	return nil
}
