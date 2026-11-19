package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"task_manager_project/pkg/constants"
	"task_manager_project/pkg/db"
)

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		errorResponse(w, "ошибка десериализации JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		errorResponse(w, "не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	if err := checkDate(&task, time.Now()); err != nil {
		errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), databaseTimeout)
	defer cancel()
	id, err := db.AddTask(ctx, &task)
	if err != nil {
		databaseErrorResponse(w, ctx, err)
		return
	}

	response := map[string]interface{}{
		"id": fmt.Sprintf("%d", id),
	}
	writeJSON(w, response)
}

func checkDate(task *db.Task, now time.Time) error {
	var rule repeatRule
	if task.Repeat != "" {
		var err error
		rule, err = parseRepeat(task.Repeat)
		if err != nil {
			return err
		}
	}

	if task.Date == "" {
		task.Date = now.Format(constants.DateFormat)
		return nil
	}

	t, err := time.Parse(constants.DateFormat, task.Date)
	if err != nil {
		return err
	}

	// Проверяем, нужно ли корректировать дату
	if !afterNow(now, t) {
		return nil // дата корректна, ничего не меняем
	}

	if task.Repeat == "" {
		task.Date = now.Format(constants.DateFormat)
		return nil
	}

	next, err := rule.nextDate(now, t)
	if err != nil {
		return err
	}

	task.Date = next
	return nil
}
