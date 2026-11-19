package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go_final_project/pkg/db"

	"github.com/stretchr/testify/require"
)

func TestTaskHandlerRejectsUnsupportedRepeat(t *testing.T) {
	rules := []struct {
		repeat  string
		message string
	}{
		{"w 1", "еженедельное повторение пока не поддерживается"},
		{"m 1", "ежемесячное повторение пока не поддерживается"},
		{"d 0", "интервал повторения"},
		{"d -1", "интервал повторения"},
	}
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, date := range []string{"", time.Now().Format("20060102"), "20990126"} {
			for _, rule := range rules {
				t.Run(method+"/"+date+"/"+rule.repeat, func(t *testing.T) {
					body, err := json.Marshal(db.Task{ID: "1", Title: "Тест", Date: date, Repeat: rule.repeat})
					require.NoError(t, err)
					request := httptest.NewRequest(method, "/api/task", bytes.NewReader(body))
					response := httptest.NewRecorder()
					TaskHandler(response, request)
					require.Equal(t, http.StatusBadRequest, response.Code)
					var result map[string]string
					require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
					require.Contains(t, result["error"], rule.message)
				})
			}
		}
	}
}
