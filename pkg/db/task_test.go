package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsDateQuery(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{"29.02.2024", true},
		{"26.01.2024", true},
		{"29.02.2023", false},
		{"20240126", false},
		{"Поход в бассейн", false},
		{"", false},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			require.Equal(t, test.want, isDateQuery(test.query))
		})
	}
}

func TestConvertDateFormat(t *testing.T) {
	require.Equal(t, "20240229", convertDateFormat("29.02.2024"))
}
