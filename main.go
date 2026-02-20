package main

import (
	"fmt"
	"go_final_project/pkg/api"
	"go_final_project/pkg/db"
	"log"
	"net/http"
)

func main() {
	const port = 7540
	const webDir = "./web"

	dbFile := "data/scheduler.db"
	err := db.Init(dbFile)
	if err != nil {
		log.Fatal("ошибка инициализации БД:", err)
	}

	api.Init()

	fs := http.FileServer(http.Dir(webDir))

	http.Handle("/", fs)

	log.Printf("Сервер запущен на http://localhost:%d", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
