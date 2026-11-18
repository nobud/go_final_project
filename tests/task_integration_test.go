//go:build integration

package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"task_manager_project/pkg/api"
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
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.DeleteTask(ctx, task.ID)
	})

	t.Run("canceled database operations", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := db.GetTask(ctx, task.ID)
		require.ErrorIs(t, err, context.Canceled)
		_, err = db.Tasks(ctx, "", 50)
		require.ErrorIs(t, err, context.Canceled)
		_, err = db.AddTask(ctx, &task)
		require.ErrorIs(t, err, context.Canceled)
		require.ErrorIs(t, db.UpdateTask(ctx, &task), context.Canceled)
		require.ErrorIs(t, db.UpdateDate(ctx, task.ID, "20990301"), context.Canceled)
		require.ErrorIs(t, db.DeleteTask(ctx, task.ID), context.Canceled)
	})

	t.Run("request deadline reaches database", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
		defer cancel()
		request := httptest.NewRequest(http.MethodGet, "/api/task?id="+task.ID, nil).WithContext(ctx)
		response := httptest.NewRecorder()
		api.GetTaskHandler(response, request)
		require.Equal(t, http.StatusGatewayTimeout, response.Code)
		require.JSONEq(t, `{"error":"превышено время ожидания БД"}`, response.Body.String())
	})

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.GetTask(ctx, task.ID)
	require.True(t, errors.Is(err, sql.ErrNoRows), "задача должна быть удалена: %v", err)

	t.Run("missing task returns 404", func(t *testing.T) {
		body, err := json.Marshal(task)
		require.NoError(t, err)
		tests := []struct {
			method  string
			path    string
			handler http.HandlerFunc
		}{
			{http.MethodGet, "/api/task?id=" + task.ID, api.TaskHandler},
			{http.MethodPut, "/api/task", api.TaskHandler},
			{http.MethodDelete, "/api/task?id=" + task.ID, api.TaskHandler},
			{http.MethodPost, "/api/task/done?id=" + task.ID, api.TaskDoneHandler},
		}
		for _, test := range tests {
			t.Run(test.method+test.path, func(t *testing.T) {
				request := httptest.NewRequest(test.method, test.path, strings.NewReader(string(body)))
				response := httptest.NewRecorder()
				test.handler(response, request)
				require.Equal(t, http.StatusNotFound, response.Code)
				require.JSONEq(t, `{"error":"задача не найдена"}`, response.Body.String())
			})
		}
	})
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
