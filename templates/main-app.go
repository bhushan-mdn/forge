package main

import (
	"log"
	"html/template"
	"net/http"
)

func main() {
	log.Println("starting server")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":8080", nil)
}
