package api

import (
	"context"
	"net/http"

	"task_manager_project/pkg/db"
)

const limit = 50

// структура ответа со списком задач
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// обрабатывает GET запросы для получения списка задач
// поддерживает параметр search для поиска по тексту или дате
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")

	ctx, cancel := context.WithTimeout(r.Context(), databaseTimeout)
	defer cancel()
	tasks, err := db.Tasks(ctx, search, limit)
	if err != nil {
		databaseErrorResponse(w, ctx, err)
		return
	}

	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	writeJSON(w, TasksResp{
		Tasks: tasks,
	})
}
