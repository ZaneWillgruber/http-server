package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("app/"))
	mux.Handle("/app/", http.StripPrefix("/app/", fs))
	mux.HandleFunc("/healthz", readinessEndpoint)

	server := http.Server{Handler: mux, Addr: ":8080"}

	fmt.Println("Listening locally at: http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		println(err.Error)
	}
}

func readinessEndpoint(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}
