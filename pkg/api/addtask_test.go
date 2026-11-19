package api

import (
	"testing"
	"time"

	"task_manager_project/pkg/db"

	"github.com/stretchr/testify/require"
)

func TestCheckDate(t *testing.T) {
	now := time.Date(2024, time.January, 26, 18, 30, 0, 0, time.FixedZone("MSK", 3*60*60))
	tests := []struct {
		name   string
		date   string
		repeat string
		want   string
	}{
		{"empty", "", "", "20240126"},
		{"empty repeating", "", "d 7", "20240126"},
		{"today", "20240126", "", "20240126"},
		{"today repeating", "20240126", "d 7", "20240126"},
		{"future", "20240201", "", "20240201"},
		{"future repeating", "20240201", "y", "20240201"},
		{"past", "20240101", "", "20240126"},
		{"past daily", "20240113", "d 7", "20240127"},
		{"past yearly", "20230126", "y", "20250126"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			task := db.Task{Date: test.date, Repeat: test.repeat, Title: "Тест", Comment: "Комментарий"}
			require.NoError(t, checkDate(&task, now))
			require.Equal(t, test.want, task.Date)
			require.Equal(t, test.repeat, task.Repeat)
			require.Equal(t, "Тест", task.Title)
			require.Equal(t, "Комментарий", task.Comment)
		})
	}
}

func TestCheckDateInvalidInput(t *testing.T) {
	now := time.Date(2024, time.January, 26, 0, 0, 0, 0, time.UTC)
	tests := []db.Task{
		{Date: "20240192"},
		{Date: "28.01.2024"},
		{Date: "20230229"},
		{Repeat: "d 0"},
		{Date: "20240126", Repeat: "d -1"},
		{Date: "20240201", Repeat: "ooops"},
		{Date: "20240101", Repeat: "d 401"},
	}
	for _, task := range tests {
		t.Run(task.Date+"/"+task.Repeat, func(t *testing.T) {
			require.Error(t, checkDate(&task, now))
		})
	}
}
