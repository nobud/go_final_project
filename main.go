package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go_final_project/pkg/api"
	"go_final_project/pkg/db"
)

const portDefault = 7540
const webDir = "./web"

func main() {

	port := getPort()

	err := db.Init()
	if err != nil {
		log.Fatal("ошибка инициализации БД:", err)
	}

	defer db.Close()

	api.Init()

	fs := http.FileServer(http.Dir(webDir))

	http.Handle("/", fs)

	log.Printf("Сервер запущен на http://localhost:%d", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

func getPort() int {
	portStr := os.Getenv("TODO_PORT")
	if portStr == "" {
		return portDefault
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Ошибка преобразования TODO_PORT='%s' в число, используется порт %d", portStr, port)
		return portDefault
	}

	return port
}
