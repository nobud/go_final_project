package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// Общий лимит для всех обращений к БД в одном HTTP-запросе.
const databaseTimeout = 5 * time.Second

func validTaskID(id string) bool {
	value, err := strconv.ParseInt(id, 10, 32)
	return err == nil && value > 0
}

func databaseErrorResponse(w http.ResponseWriter, ctx context.Context, err error) {
	if errors.Is(ctx.Err(), context.Canceled) {
		return
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		errorResponse(w, "превышено время ожидания БД", http.StatusGatewayTimeout)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		errorResponse(w, "задача не найдена", http.StatusNotFound)
		return
	}
	errorResponse(w, err.Error(), http.StatusInternalServerError)
}

func errorResponse(w http.ResponseWriter, errMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)

	response := map[string]string{"error": errMsg}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Ошибка сериализации JSON", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Ошибка сериализации JSON", http.StatusInternalServerError)
	}
}
