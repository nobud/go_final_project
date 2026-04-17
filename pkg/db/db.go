package db

import (
	"database/sql"
	"fmt"
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
	port := "5432"
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "admin")
	dbname := getEnv("DB_NAME", "scheduler")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("ошибка открытия подключения: %w", err)
	}

	// Создаем таблицу если её нет
	if err = createTableIfNotExists(); err != nil {
		db.Close()
		return fmt.Errorf("ошибка создания таблицы: %w", err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createTableIfNotExists() error {
	var exists bool
	queryCheck := `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_name = 'scheduler'
        )`

	err := db.QueryRow(queryCheck).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования таблицы: %w", err)
	}

	if !exists {
		_, err = db.Exec(schemaPostgres)
		if err != nil {
			return fmt.Errorf("ошибка создания таблицы: %w", err)
		}
	}

	return nil
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
