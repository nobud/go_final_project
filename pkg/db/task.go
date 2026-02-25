package db

import (
	"database/sql"
	"fmt"
	"time"

	"go_final_project/pkg/constants"
)

const (
	inputDateFormat = `02.01.2006`
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func UpdateDate(id string, newDate string) error {
	query := `UPDATE scheduler SET date = $1 WHERE id = $2`

	result, err := db.Exec(query, newDate, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления даты задачи: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("задача с ID = %s не найдена", id)
	}

	return nil
}

func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = $1`

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления задачи: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("задача с ID = %s не найдена", id)
	}

	return nil

}

func AddTask(task *Task) (int64, error) {
	query := `
        INSERT INTO scheduler (date, title, comment, repeat)
        VALUES ($1, $2, $3, $4) RETURNING id`

	var id int64
	err := db.QueryRow(query, task.Date, task.Title, task.Comment, task.Repeat).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления задачи: %w", err)
	}

	return id, nil
}

func GetTask(id string) (*Task, error) {
	var task Task
	query := `
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = $1`

	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задачи с ID = %s: %w", id, err)
	}
	return &task, nil
}

func UpdateTask(task *Task) error {
	query := `
        UPDATE scheduler SET date = $1, title = $2, comment = $3, repeat = $4
        WHERE id = $5`

	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления задачи: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("задача с ID = %s не найдена", task.ID)
	}

	return nil
}

func Tasks(search string, limit int) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	switch {
	case isDateQuery(search):
		date := convertDateFormat(search)
		query := `
        SELECT id, date, title, comment, repeat
        FROM scheduler
        WHERE date = $1
        ORDER BY date
        LIMIT $2`
		rows, err = db.Query(query, date, limit)

	case search != "":
		pattern := "%" + search + "%"
		query := `
        SELECT id, date, title, comment, repeat
        FROM scheduler
        WHERE title LIKE $1 OR comment LIKE $2
        ORDER BY date
        LIMIT $3`
		rows, err = db.Query(query, pattern, pattern, limit)

	default:
		query := `
        SELECT id, date, title, comment, repeat
        FROM scheduler
        ORDER BY date
        LIMIT $1`
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
	_, err := time.Parse(inputDateFormat, s)
	return err == nil
}

func convertDateFormat(dateStr string) string {
	t, _ := time.Parse(inputDateFormat, dateStr)
	return t.Format(constants.DateFormat)
}
