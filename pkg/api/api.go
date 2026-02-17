package api

import (
	"net/http"
)

const DateFormat = "20060102"

func Init() {

	http.HandleFunc("/api/nextdate", NextDateHandler)

	http.HandleFunc("/api/task", TaskHandler)
}
