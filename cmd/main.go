package main

import (
	"app/config"
	"app/handler"
	"app/internal/middleware"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file if it exists (optional for Docker environments)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	config.ConnectDB()
	port := os.Getenv("PORT")
	serverAddress := fmt.Sprintf(":%s", port)
	NewServer := chi.NewRouter()
	NewServer.Use(middleware.Logger)
	handler.GetHandlers(NewServer)
	log.Fatal(http.ListenAndServe(serverAddress, NewServer))
	log.Println("Server starting")
}
