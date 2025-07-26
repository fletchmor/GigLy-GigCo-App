<<<<<<< HEAD
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
	config.ConnectDB()
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	port := os.Getenv("PORT")
	serverAddress := fmt.Sprintf(":%s", port)
	NewServer := chi.NewRouter()
	NewServer.Use(middleware.Logger)
	handler.GetHandlers(NewServer)
	log.Fatal(http.ListenAndServe(serverAddress, NewServer))
	log.Println("Server starting")
}
=======
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
>>>>>>> 86c55890c09c3f69b573f338c0a66df1e5fdb519
