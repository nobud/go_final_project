package db

import (
	"database/sql"
	"fmt"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	query := `
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (?, ?, ?, ?)`

	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления задачи: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID: %w", err)
	}

	return id, nil
}

func Tasks(search string, limit int) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	// проверка является ли search датой
	if isDateQuery(search) {
		// поиск по дате
		// преобразование даты в формат 20060102
		date := convertDateFormat(search)
		query := `
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE date = ?
		ORDER BY date
		LIMIT ?`
		rows, err = db.Query(query, date, limit)
	} else if search != "" {
		// поиск по подстроке в title или в comment
		pattern := "%" + search + "%"
		query := `
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE title LIKE ? OR comment LIKE ?
		ORDER BY date
		LIMIT ?`
		rows, err = db.Query(query, pattern, pattern, limit)
	} else {
		// без поиска - все задачи
		query := `
		SELECT id, date, title, comment, repeat
		FROM scheduler
		ORDER BY date
		LIMIT ?`
		rows, err = db.Query(query, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка задач: %w", err)
	}

	defer rows.Close()
	tasks := make([]*Task, 0)

	for rows.Next() {
		task := Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования задачи: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов: %w", err)
	}

	return tasks, nil
}

func isDateQuery(s string) bool {
	if len(s) != 10 {
		return false
	}
	if s[2] != '.' || s[5] != '.' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if i == 2 || i == 5 {
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func convertDateFormat(dateStr string) string {
	day := dateStr[0:2]
	month := dateStr[3:5]
	year := dateStr[6:10]
	return year + month + day
}
