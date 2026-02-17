package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	const port = 7540
	const webDir = "./web"

	fs := http.FileServer(http.Dir(webDir))

	http.Handle("/", fs)

	log.Printf("Сервер запущен на http://localhost:%d", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}

}
