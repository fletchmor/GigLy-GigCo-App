package main

import (
	"app/handler"
	"app/internal/middleware"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	NewServer := chi.NewRouter()
	NewServer.Use(middleware.Logger)
	handler.GetHandlers(NewServer)
	log.Fatal(http.ListenAndServe(":8080", NewServer))
	log.Println("Server starting on port :8080")
}
