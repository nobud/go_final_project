package api

import (
	"go_final_project/pkg/db"
	"net/http"
)

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
	const limit = 50

	tasks, err := db.Tasks(search, limit)
	if err != nil {
		errorResponse(w, "ошибка получения задач "+err.Error(), http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	writeJSON(w, TasksResp{
		Tasks: tasks,
	})
}
