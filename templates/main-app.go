package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("starting server")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>app</title>
		</head>
		<body>
			<p>hello world</p>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	})

	http.ListenAndServe(":8080", nil)
}
