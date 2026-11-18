package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"task_manager_project/pkg/api"
	"task_manager_project/pkg/db"
)

const portDefault = 7540
const webDir = "./web"
const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	port := getPort()

	err := db.Init(ctx)
	if err != nil {
		return fmt.Errorf("ошибка инициализации БД: %w", err)
	}

	defer db.Close()

	api.Init()

	fs := http.FileServer(http.Dir(webDir))

	http.Handle("/", fs)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           http.DefaultServeMux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("ошибка запуска сервера: %w", err)
	}
	log.Printf("Сервер запущен на http://localhost:%d", port)
	return serveHTTP(ctx, server, listener)
}

func serveHTTP(ctx context.Context, server *http.Server, listener net.Listener) error {
	// При превышении времени Shutdown закрываем оставшиеся соединения.
	defer server.Close()
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("ошибка HTTP-сервера: %w", err)
	case <-ctx.Done():
		log.Print("Остановка HTTP-сервера")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("ошибка остановки HTTP-сервера: %w", err)
		}
		return nil
	}
}

func getPort() int {
	portStr := os.Getenv("TODO_PORT")
	if portStr == "" {
		return portDefault
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		log.Printf("Некорректный TODO_PORT=%q, ожидается число от 1 до 65535, используется порт %d", portStr, portDefault)
		return portDefault
	}

	return port
}
