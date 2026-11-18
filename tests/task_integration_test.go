//go:build integration

package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"task_manager_project/pkg/db"

	"github.com/stretchr/testify/require"
)

func TestTaskCRUD(t *testing.T) {
	task := db.Task{
		Date: "20990126", Title: "Подготовить отчёт",
		Comment: "Интеграционный тест", Repeat: "d 7",
	}
	var created struct {
		ID string `json:"id"`
	}
	taskRequest(t, http.MethodPost, "/api/task", task, &created)
	require.NotEmpty(t, created.ID)
	task.ID = created.ID
	// Удаляем задачу и при досрочном завершении проверки.
	t.Cleanup(func() { _ = db.DeleteTask(task.ID) })

	var fetched db.Task
	taskRequest(t, http.MethodGet, "/api/task?id="+task.ID, nil, &fetched)
	require.Equal(t, task, fetched)

	task.Title = "Обновлённый отчёт"
	task.Date = "20990201"
	task.Comment = ""
	task.Repeat = ""
	var result map[string]string
	taskRequest(t, http.MethodPut, "/api/task", task, &result)
	require.Empty(t, result)
	taskRequest(t, http.MethodGet, "/api/task?id="+task.ID, nil, &fetched)
	require.Equal(t, task, fetched)

	taskRequest(t, http.MethodDelete, "/api/task?id="+task.ID, nil, &result)
	require.Empty(t, result)
	_, err := db.GetTask(task.ID)
	require.True(t, errors.Is(err, sql.ErrNoRows), "задача должна быть удалена: %v", err)
}

func taskRequest(t *testing.T, method, path string, body any, result any) {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)
	request, err := http.NewRequest(method, testServerURL+path, bytes.NewReader(data))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NoError(t, json.NewDecoder(response.Body).Decode(result))
}
