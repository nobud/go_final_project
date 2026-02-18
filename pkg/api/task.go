package api

import (
	"encoding/json"
	"go_final_project/pkg/db"
	"net/http"
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
	default:
		errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		errorResponse(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		errorResponse(w, err.Error(), http.StatusBadRequest)
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

	if task.ID == "" {
		errorResponse(w, "не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		errorResponse(w, "не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.UpdateTask(&task)
	if err != nil {
		errorResponse(w, "ошибка обновления задачи в БД: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, struct{}{})
}
