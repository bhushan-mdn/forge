package main

import (
	"log"
	"net/http"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xw := NewXResponseWriter(w)

		next.ServeHTTP(xw, r)

		log.Println(xw.statusCode, http.StatusText(xw.statusCode), r.URL.String())
	})
}

type XResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewXResponseWriter(w http.ResponseWriter) *XResponseWriter {
	return &XResponseWriter{w, http.StatusOK}
}

func (xw *XResponseWriter) WriteHeader(code int) {
	xw.statusCode = code
	xw.ResponseWriter.WriteHeader(code)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Encoding", "application/json")
		w.Write([]byte(`{"message": "hello world"}`))
	})
	mux.HandleFunc("/bad", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Encoding", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "hello world"}`))
	})
	mux.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Encoding", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "hello world"}`))
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: loggingMiddleware(mux),
	}

	log.Println("starting server on port :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Printf("error: %+v\n", err)
	}
}
