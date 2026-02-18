package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// afterNow возвращает true, если date > now (сравнивает только даты, без времени)
func afterNow(date, now time.Time) bool {
	dateStr := date.Format(DateFormat)
	nowStr := now.Format(DateFormat)
	return dateStr > nowStr
}

func newDay(now time.Time, start time.Time, parts []string) (string, error) {
	maxInterval := 400
	// проверка формата
	if len(parts) != 2 {
		return "", fmt.Errorf("Некорректный формат правила повторения: %s", strings.Join(parts, " "))
	}

	// конвертация интервала
	interval, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("некорректный интервал: %s", parts[1])
	}

	// проверка максимального интервала
	if interval > maxInterval {
		return "", fmt.Errorf("превышен максимально допустимый интервал (%d):", maxInterval, interval)
	}

	// поиск следующей даты
	date := start
	for {
		date = date.AddDate(0, 0, interval)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

func newYear(now time.Time, start time.Time, parts []string) (string, error) {
	// проверка формата
	if len(parts) != 1 {
		return "", fmt.Errorf("Некорректный формат правила повторения: %s", strings.Join(parts, " "))
	}

	// поиск следующей даты
	date := start
	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	//проверка на пустое правило повторения
	if repeat == "" {
		return "", fmt.Errorf("пустое правило повторения")
	}

	// проверка на неправильный формат исходной даты
	start, err := time.Parse("20060102", dstart)
	if err != nil {
		return "", fmt.Errorf("некорректный формат даты: %s", dstart)
	}

	// парсинг правила повторения
	parts := strings.Split(repeat, " ")
	if len(parts) == 0 {
		return "", fmt.Errorf("пустое правило: %s", repeat)
	}

	switch parts[0] {
	case "d":
		return newDay(now, start, parts)
	case "y":
		return newYear(now, start, parts)
	case "m", "w":
		return "", fmt.Errorf("Неподдерживаемый формат правила: %s", repeat)
	default:
		return "", fmt.Errorf("Некорректный формат правила: %s", repeat)
	}
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")

	if dateParam == "" {
		errorResponse(w, "Не указан параметр date", http.StatusBadRequest)
		return
	}

	if repeatParam == "" {
		errorResponse(w, "Не указан параметр repeat", http.StatusBadRequest)
		return
	}

	var now time.Time
	if nowParam == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateFormat, nowParam)
		if err != nil {
			errorResponse(w, "Неверный формат параметра now: "+nowParam, http.StatusBadRequest)
			return
		}
	}

	next, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(next))
}
