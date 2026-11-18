//go:build integration

package tests

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"task_manager_project/pkg/api"
	"task_manager_project/pkg/db"
)

var testServerURL string

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestMain(m *testing.M) {
	code, err := runTests(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(code)
}

func runTests(m *testing.M) (int, error) {
	// Локально создаём временную базу. В CI адрес базы приходит из окружения.
	if os.Getenv("DB_HOST") == "" {
		containerName := "todo-test-" + strings.ToLower(rand.Text())
		// defer выполнит очистку и при ошибке подготовки тестов.
		defer func() {
			if _, err := dockerCommand("rm", "--force", containerName); err != nil {
				fmt.Fprintln(os.Stderr, "Очистка временного контейнера:", err)
			}
		}()
		if err := startPostgres(containerName); err != nil {
			return 1, err
		}
	}

	// Тесты очищают таблицу, поэтому разрешаем только тестовую базу.
	if os.Getenv("DB_NAME") != "scheduler_test" {
		return 1, fmt.Errorf("для тестов укажите DB_NAME=scheduler_test: таблица задач будет очищена")
	}
	if err := waitForPostgres(); err != nil {
		return 1, err
	}
	if err := db.Init(context.Background()); err != nil {
		return 1, err
	}
	defer db.Close()

	// Поднимаем HTTP-сервер на свободном порту для запросов из тестов.
	api.Init()
	http.Handle("/", http.FileServer(http.Dir("../web")))
	server := httptest.NewServer(http.DefaultServeMux)
	defer server.Close()
	testServerURL = server.URL
	// После m.Run() отработают defer: сервер, подключение и контейнер закроются.
	return m.Run(), nil
}

func startPostgres(containerName string) error {
	password := rand.Text()
	_, err := dockerCommand("run",
		"--detach", "--rm", "--name", containerName,
		"--publish", "127.0.0.1::5432", // Docker сам выберет свободный порт.
		"--tmpfs", "/var/lib/postgresql", // Данные хранятся в памяти.
		"--env", "POSTGRES_DB=scheduler_test",
		"--env", "POSTGRES_USER=postgres",
		"--env", "POSTGRES_PASSWORD="+password,
		"postgres:18-alpine",
	)
	if err != nil {
		return fmt.Errorf("для тестов запустите Docker: %w", err)
	}

	address, err := dockerCommand("port", containerName, "5432/tcp")
	if err != nil {
		return err
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("адрес временной PostgreSQL: %w", err)
	}

	// Приложение и тесты будут подключаться к одной временной базе.
	environment := map[string]string{
		"DB_HOST":     host,
		"DB_PORT":     port,
		"DB_NAME":     "scheduler_test",
		"DB_USER":     "postgres",
		"DB_PASSWORD": password,
	}
	for key, value := range environment {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Контейнер может запуститься раньше, чем PostgreSQL начнёт принимать запросы.
func waitForPostgres() error {
	connURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(getEnvOrDefault("DB_USER", "postgres"), os.Getenv("DB_PASSWORD")),
		Host:     net.JoinHostPort(os.Getenv("DB_HOST"), getEnvOrDefault("DB_PORT", "5432")),
		Path:     "/scheduler_test",
		RawQuery: "sslmode=disable",
	}
	connection, err := sql.Open("pgx", connURL.String())
	if err != nil {
		return err
	}
	defer connection.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for ctx.Err() == nil {
		err = connection.PingContext(ctx)
		if err == nil {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("PostgreSQL не готова к тестам: %w", err)
}

// Ограничиваем время команд Docker, чтобы тесты не зависли при сбое.
func dockerCommand(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	output, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker %s: %w\n%s", args[0], err, output)
	}
	return strings.TrimSpace(string(output)), nil
}
