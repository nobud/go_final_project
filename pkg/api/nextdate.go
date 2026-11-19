package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"task_manager_project/pkg/constants"
)

// afterNow возвращает true, если date > now (сравнивает только даты, без времени)
func afterNow(date, now time.Time) bool {
	dateStr := date.Format(constants.DateFormat)
	nowStr := now.Format(constants.DateFormat)
	return dateStr > nowStr
}

type repeatRule struct {
	years int
	days  int
}

func parseRepeat(repeat string) (repeatRule, error) {
	const maxInterval = 400
	if repeat == "" {
		return repeatRule{}, fmt.Errorf("пустое правило повторения")
	}
	parts := strings.Split(repeat, " ")
	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			break
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > maxInterval {
			return repeatRule{}, fmt.Errorf("интервал повторения должен быть целым числом от 1 до %d: %s", maxInterval, parts[1])
		}
		return repeatRule{days: days}, nil
	case "y":
		if len(parts) == 1 {
			return repeatRule{years: 1}, nil
		}
	case "w":
		return repeatRule{}, fmt.Errorf("еженедельное повторение пока не поддерживается")
	case "m":
		return repeatRule{}, fmt.Errorf("ежемесячное повторение пока не поддерживается")
	}
	return repeatRule{}, fmt.Errorf("некорректный формат правила повторения: %s", repeat)
}

func (rule repeatRule) nextDate(now, date time.Time) (string, error) {
	for {
		date = date.AddDate(rule.years, 0, rule.days)
		if date.Year() > 9999 {
			return "", fmt.Errorf("следующая дата выходит за пределы поддерживаемого формата")
		}
		if afterNow(date, now) {
			return date.Format(constants.DateFormat), nil
		}
	}
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	rule, err := parseRepeat(repeat)
	if err != nil {
		return "", err
	}

	start, err := time.Parse(constants.DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректный формат даты: %s", dstart)
	}
	return rule.nextDate(now, start)
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
		now, err = time.Parse(constants.DateFormat, nowParam)
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
	if _, err := w.Write([]byte(next)); err != nil {
		log.Printf("Response write error: %v", err)
	}
}
