package main

import (
	"app/config"
	_ "app/docs"
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

// @title GigCo API
// @version 1.0
// @description GigCo platform API for gig workers and job management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load .env file if it exists (optional for Docker environments)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	// Initialize database
	config.ConnectDB()

	// Initialize JWT
	auth.InitJWT()

	// Initialize payment configuration (optional - warnings only if not configured)
	config.InitPaymentConfig()

	port := os.Getenv("PORT")
	serverAddress := fmt.Sprintf(":%s", port)
	NewServer := chi.NewRouter()
	NewServer.Use(middleware.Logger)
	
	// Public routes (no JWT required)
	handler.GetPublicHandlers(NewServer)
	handler.PostPublicHandlers(NewServer)
	
	// Protected routes (JWT required)
	NewServer.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth)
		handler.GetHandlers(r)
		handler.PostHandlers(r)
		handler.PutHandlers(r)
		handler.DeleteHandlers(r)
	})
	
	log.Println("Server starting")
	log.Fatal(http.ListenAndServe(serverAddress, NewServer))
}
