package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // драйвер pgx в режиме совместимости с database/sql
)

var db *sql.DB

const schemaPostgres = `
    CREATE TABLE IF NOT EXISTS scheduler (
        id SERIAL PRIMARY KEY,
        date VARCHAR(8) NOT NULL DEFAULT '',
        title VARCHAR(255) NOT NULL DEFAULT '',
        comment TEXT,
        repeat VARCHAR(128) NOT NULL DEFAULT ''
    );
    CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init инициализирует подключение к PostgreSQL
func Init() error {
	// DB_HOST можно задать через окружение, по умолчанию localhost
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return errors.New("укажите пароль PostgreSQL в DB_PASSWORD")
	}
	dbname := getEnv("DB_NAME", "scheduler")

	connURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + dbname,
		RawQuery: "sslmode=disable",
	}

	var err error
	db, err = sql.Open("pgx", connURL.String())
	if err != nil {
		return fmt.Errorf("ошибка открытия подключения: %w", err)
	}

	if _, err = db.Exec(schemaPostgres); err != nil {
		db.Close()
		return fmt.Errorf("ошибка инициализации схемы: %w", err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
