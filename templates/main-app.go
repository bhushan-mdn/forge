package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := \`
		<!DOCTYPE html>
		<html>
		<head>
			<title>App</title>
			<style>body { font-family: sans-serif; text-align: center; padding-top: 50px; }</style>
		</head>
		<body>
			<h1>Hello from your Go Web App!</h1>
			<p>Server started on port 8080.</p>
		</body>
		</html>
		\`
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	})

	log.Println("🌐 Starting web app on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("error: %v", err)
	}
}
