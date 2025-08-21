package main

import (
	"app/config"
	"app/handler"
	"app/internal/auth"
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

	// Initialize JWT
	auth.InitJWT()

	port := os.Getenv("PORT")
	serverAddress := fmt.Sprintf(":%s", port)
	NewServer := chi.NewRouter()
	NewServer.Use(middleware.Logger)
	NewServer.Use(middleware.JWTAuth)
	handler.GetHandlers(NewServer)
	handler.PostHandlers(NewServer)
	handler.PutHandlers(NewServer)
	handler.DeleteHandlers(NewServer)
	log.Println("Server starting")
	log.Fatal(http.ListenAndServe(serverAddress, NewServer))
}
