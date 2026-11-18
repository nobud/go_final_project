package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Эти запросы должны отклоняться до обращения к базе данных.
func TestHandlersRejectInvalidRequests(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    string
		status  int
	}{
		{"post invalid JSON", TaskHandler, "POST", "/api/task", "{", 400},
		{"post missing title", TaskHandler, "POST", "/api/task", `{}`, 400},
		{"post invalid date", TaskHandler, "POST", "/api/task", `{"title":"Тест","date":"20240192"}`, 400},
		{"post invalid repeat", TaskHandler, "POST", "/api/task", `{"title":"Тест","repeat":"d 0"}`, 400},
		{"put invalid JSON", TaskHandler, "PUT", "/api/task", "{", 400},
		{"put missing ID", TaskHandler, "PUT", "/api/task", `{"title":"Тест"}`, 400},
		{"put missing title", TaskHandler, "PUT", "/api/task", `{"id":"1"}`, 400},
		{"put invalid repeat", TaskHandler, "PUT", "/api/task", `{"id":"1","title":"Тест","repeat":"d -1"}`, 400},
		{"get missing ID", TaskHandler, "GET", "/api/task", "", 400},
		{"delete missing ID", TaskHandler, "DELETE", "/api/task", "", 400},
		{"task unsupported method", TaskHandler, "PATCH", "/api/task", "", 405},
		{"tasks unsupported method", TasksHandler, "POST", "/api/tasks", "", 405},
		{"done missing ID", TaskDoneHandler, "POST", "/api/task/done", "", 400},
		{"done unsupported method", TaskDoneHandler, "GET", "/api/task/done", "", 405},
		{"nextdate missing date", NextDateHandler, "GET", "/api/nextdate?repeat=y", "", 400},
		{"nextdate missing repeat", NextDateHandler, "GET", "/api/nextdate?date=20240126", "", 400},
		{"nextdate invalid now", NextDateHandler, "GET", "/api/nextdate?now=bad&date=20240126&repeat=y", "", 400},
		{"nextdate invalid date", NextDateHandler, "GET", "/api/nextdate?date=bad&repeat=y", "", 400},
		{"nextdate invalid repeat", NextDateHandler, "GET", "/api/nextdate?date=20240126&repeat=d%200", "", 400},
		{"nextdate unsupported method", NextDateHandler, "POST", "/api/nextdate", "", 405},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			response := httptest.NewRecorder()
			test.handler(response, request)
			require.Equal(t, test.status, response.Code)
			require.Contains(t, response.Header().Get("Content-Type"), "application/json")
			var result map[string]string
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
			require.NotEmpty(t, result["error"])
		})
	}
}

func TestNextDateHandler(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/nextdate?now=20240126&date=20240113&repeat=d%207", nil)
	response := httptest.NewRecorder()
	NextDateHandler(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Header().Get("Content-Type"), "text/plain")
	require.Equal(t, "20240127", response.Body.String())
}
