package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNextDate(t *testing.T) {
	now := time.Date(2024, time.January, 26, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		date   string
		repeat string
		want   string
	}{
		{"16890220", "y", "20240220"},
		{"20250701", "y", "20260701"},
		{"20240101", "y", "20250101"},
		{"20231231", "y", "20241231"},
		{"20240229", "y", "20250301"},
		{"20240301", "y", "20250301"},
		{"20240113", "d 7", "20240127"},
		{"20240120", "d 20", "20240209"},
		{"20240202", "d 30", "20240303"},
		{"20231225", "d 12", "20240130"},
		{"20240228", "d 1", "20240229"},
		{"20240126", "d 1", "20240127"},
		{"20240126", "d 400", "20250301"},
	}
	for _, test := range tests {
		t.Run(test.date+"/"+test.repeat, func(t *testing.T) {
			date, err := NextDate(now, test.date, test.repeat)
			require.NoError(t, err)
			require.Equal(t, test.want, date)
		})
	}
}

func TestNextDateInvalidInput(t *testing.T) {
	now := time.Date(2024, time.January, 26, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		date   string
		repeat string
	}{
		{"20240126", ""},
		{"20240126", "k 34"},
		{"20240126", "ooops"},
		{"15000156", "y"},
		{"ooops", "y"},
		{"20230229", "y"},
		{"20240113", "d"},
		{"20240113", "d 0"},
		{"20240113", "d -1"},
		{"20240113", "d 401"},
		{"20240113", "d abc"},
		{"20240113", "d 7 extra"},
		{"20240113", "y 2"},
		{"20240113", "w 1"},
		{"20240113", "m 1"},
		{"99991231", "d 1"},
		{"99991231", "y"},
	}
	for _, test := range tests {
		t.Run(test.date+"/"+test.repeat, func(t *testing.T) {
			date, err := NextDate(now, test.date, test.repeat)
			require.Error(t, err)
			require.Empty(t, date)
		})
	}
}
