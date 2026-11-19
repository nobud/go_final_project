package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetPort(t *testing.T) {
	tests := []struct {
		value string
		want  int
	}{
		{"", portDefault},
		{"8080", 8080},
		{"1", 1},
		{"65535", 65535},
		{"0", portDefault},
		{"-1", portDefault},
		{"65536", portDefault},
		{"abc", portDefault},
		{"999999999999999999999999", portDefault},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			t.Setenv("TODO_PORT", test.value)
			require.Equal(t, test.want, getPort())
		})
	}
}

func TestServeHTTPGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	requestStarted := make(chan struct{})
	finishRequest := make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		select {
		case <-finishRequest:
			_, _ = w.Write([]byte("done"))
		case <-r.Context().Done():
		}
	})}
	t.Cleanup(func() { server.Close() })
	serverResult := make(chan error, 1)
	go func() { serverResult <- serveHTTP(ctx, server, listener) }()
	requestResult := make(chan error, 1)
	responseBody := make(chan string, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://" + listener.Addr().String())
		if err != nil {
			requestResult <- err
			return
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		responseBody <- string(body)
		requestResult <- err
	}()
	select {
	case <-requestStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("HTTP-запрос не дошёл до обработчика")
	}

	cancel()
	select {
	case err := <-serverResult:
		t.Fatalf("сервер завершился до окончания запроса: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(finishRequest)
	select {
	case err := <-requestResult:
		require.NoError(t, err)
		require.Equal(t, "done", <-responseBody)
	case <-time.After(3 * time.Second):
		t.Fatal("сервер не завершил текущий запрос")
	}
	select {
	case err := <-serverResult:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("сервер не остановился")
	}
}
