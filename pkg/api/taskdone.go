package api

import (
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	if task.Repeat == "" { //одноразовая задача
		err = db.DeleteTask(id)
		if err != nil {
			errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else { //периодическая задача
		now := time.Now()
		// получаем следующую дату выполнения
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		//обновление даты задачи
		err = db.UpdateDate(id, next)
		if err != nil {
			errorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, struct{}{})
}
