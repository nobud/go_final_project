package api

import (
	"context"
	"net/http"
	"time"

	"task_manager_project/pkg/db"
)

func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	if task.Repeat == "" { //одноразовая задача
		err = db.DeleteTask(ctx, id)
		if err != nil {
			databaseErrorResponse(w, ctx, err)
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
		err = db.UpdateDate(ctx, id, next)
		if err != nil {
			databaseErrorResponse(w, ctx, err)
			return
		}
	}

	writeJSON(w, struct{}{})
}
