package main

import (
	"log"
	"net/http"
)

// XResponseWriter and loggingMiddleware provide basic logging for the API.
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
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xw := NewXResponseWriter(w)
		next.ServeHTTP(xw, r)
		log.Println(xw.statusCode, http.StatusText(xw.statusCode), r.URL.String())
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\": \"UP\"}"))
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: loggingMiddleware(mux),
	}

	log.Println("🚀 Starting API server on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Printf("error: %+v\n", err)
	}
}
