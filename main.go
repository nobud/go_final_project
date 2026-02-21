package main

import (
	"fmt"
	"log"
	"net/http"

	"go_final_project/pkg/api"
	"go_final_project/pkg/db"
)

const port = 7540
const webDir = "./web"

func main() {

	dbFile := "data/scheduler.db"
	err := db.Init(dbFile)
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
