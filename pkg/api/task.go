package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"task_manager_project/pkg/db"
)

// обрабатывает все запросы к /api/task
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		AddTaskHandler(w, r)
	case http.MethodPut:
		UpdateTaskHandler(w, r)
	case http.MethodGet:
		GetTaskHandler(w, r)
	case http.MethodDelete:
		DeleteTaskHandler(w, r)
	default:
		errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if !validTaskID(id) {
		errorResponse(w, "Некорректный идентификатор задачи", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), databaseTimeout)
	defer cancel()
	err := db.DeleteTask(ctx, id)
	if err != nil {
		databaseErrorResponse(w, ctx, err)
		return
	}

	writeJSON(w, struct{}{})
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if !validTaskID(id) {
		errorResponse(w, "Некорректный идентификатор задачи", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), databaseTimeout)
	defer cancel()
	task, err := db.GetTask(ctx, id)
	if err != nil {
		databaseErrorResponse(w, ctx, err)
		return
	}

	writeJSON(w, task)
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		errorResponse(w, "ошибка десериализации JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !validTaskID(task.ID) {
		errorResponse(w, "некорректный идентификатор задачи", http.StatusBadRequest)
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
	err := db.UpdateTask(ctx, &task)
	if err != nil {
		databaseErrorResponse(w, ctx, err)
		return
	}

	writeJSON(w, struct{}{})
}
