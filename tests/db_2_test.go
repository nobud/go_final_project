package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type Task struct {
	ID      int64  `db:"id"`
	Date    string `db:"date"`
	Title   string `db:"title"`
	Comment string `db:"comment"`
	Repeat  string `db:"repeat"`
}

func count(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(&count, `SELECT COUNT(id) FROM scheduler`)
	return count, err
}

// Функция для чтения переменных окружения
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func openDB(t *testing.T) *sqlx.DB {

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := "5432"
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "admin")
	dbname := getEnvOrDefault("DB_NAME", "scheduler")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sqlx.Connect("pgx", connStr)
	assert.NoError(t, err, "Не удалось подключиться к БД")

	// Очищаем таблицу перед тестом для изоляции
	_, err = db.Exec(`TRUNCATE TABLE scheduler RESTART IDENTITY CASCADE`)
	assert.NoError(t, err, "Не удалось очистить таблицу")

	return db
}

func TestDB(t *testing.T) {

	db := openDB(t)
	defer db.Close()

	// Проверяем, что таблица пуста
	before, err := count(db)
	assert.NoError(t, err)
	assert.Equal(t, 0, before, "Таблица должна быть пустой перед тестом")

	today := time.Now().Format("20060102")

	var id int64
	err = db.Get(&id, `
		INSERT INTO scheduler (date, title, comment, repeat) 
		VALUES ($1, 'Todo', 'Комментарий', '')
		RETURNING id
	`, today)
	assert.NoError(t, err, "Ошибка при вставке задачи")
	assert.Greater(t, id, int64(0), "ID должен быть положительным")

	// Проверяем, что задача создана корректно
	var task Task
	err = db.Get(&task, `SELECT * FROM scheduler WHERE id = $1`, id)
	assert.NoError(t, err, "Ошибка при получении задачи")

	assert.Equal(t, id, task.ID)
	assert.Equal(t, "Todo", task.Title)
	assert.Equal(t, "Комментарий", task.Comment)
	assert.Equal(t, today, task.Date)
	assert.Empty(t, task.Repeat)

	// Удаляем задачу
	_, err = db.Exec(`DELETE FROM scheduler WHERE id = $1`, id)
	assert.NoError(t, err, "Ошибка при удалении задачи")

	// Проверяем, что задача действительно удалена
	err = db.Get(&task, `SELECT * FROM scheduler WHERE id = $1`, id)
	assert.Error(t, err, "Задача должна быть удалена")

	// Проверяем, что количество записей вернулось к исходному
	after, err := count(db)
	assert.NoError(t, err)
	assert.Equal(t, before, after, "Количество записей должно вернуться к исходному")
}
