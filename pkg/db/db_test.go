package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitEmptyExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scheduler.db")
	require.NoError(t, os.WriteFile(path, nil, 0600))
	require.NoError(t, Init(path))
	t.Cleanup(func() { Close() })

	tasks, err := Tasks("", 50)
	require.NoError(t, err)
	require.Empty(t, tasks)
}

func TestInitPreservesTasksAndRestoresIndex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scheduler.db")
	require.NoError(t, Init(path))
	t.Cleanup(func() { Close() })
	_, err := AddTask(&Task{Date: "20990126", Title: "Сохранённая задача"})
	require.NoError(t, err)
	_, err = db.Exec("DROP INDEX idx_scheduler_date")
	require.NoError(t, err)
	require.NoError(t, Close())

	require.NoError(t, Init(path))
	tasks, err := Tasks("", 50)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.Equal(t, "Сохранённая задача", tasks[0].Title)
	var indexCount int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_scheduler_date'").Scan(&indexCount))
	require.Equal(t, 1, indexCount)
}
