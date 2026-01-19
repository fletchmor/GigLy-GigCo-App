package main

import (
	"app/config"
	_ "app/docs"
	"app/handler"
	"app/internal/auth"
	"app/internal/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Validate required configuration in production
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "production" {
		validateProductionConfig()
	}

	// Initialize database
	config.ConnectDB()

	// Initialize JWT
	auth.InitJWT()

	// Initialize payment configuration (optional - warnings only if not configured)
	config.InitPaymentConfig()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddress := fmt.Sprintf(":%s", port)

	// Initialize rate limiters
	standardLimiter := middleware.StandardRateLimit()

	// Create router
	router := chi.NewRouter()

	// Apply global middleware (order matters!)
	router.Use(middleware.SecurityHeaders)                           // Security headers first
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))      // CORS handling
	router.Use(middleware.RateLimit(standardLimiter))                // Rate limiting
	router.Use(middleware.Logger)                                    // Request logging

	// Public routes (no JWT required)
	handler.GetPublicHandlers(router)
	handler.PostPublicHandlers(router)

	// Protected routes (JWT required)
	router.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth)
		handler.GetHandlers(r)
		handler.PostHandlers(r)
		handler.PutHandlers(r)
		handler.DeleteHandlers(r)
	})

	// Configure HTTP server with timeouts
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for shutdown signals
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Graceful shutdown goroutine
	go func() {
		<-quit
		log.Println("Server is shutting down...")

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	// Check if TLS certificates are provided
	tlsCert := os.Getenv("TLS_CERT")
	tlsKey := os.Getenv("TLS_KEY")

	if tlsCert != "" && tlsKey != "" {
		log.Printf("Starting HTTPS server on %s", serverAddress)
		log.Printf("Using TLS certificate: %s", tlsCert)
		if err := server.ListenAndServeTLS(tlsCert, tlsKey); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v\n", err)
		}
	} else {
		log.Printf("Starting HTTP server on %s", serverAddress)
		if appEnv == "production" {
			log.Println("WARNING: Running without TLS in production - this is not recommended!")
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v\n", err)
		}
	}

	<-done
	log.Println("Server stopped gracefully")
}

// validateProductionConfig ensures required configuration is set for production
func validateProductionConfig() {
	required := []string{
		"JWT_SECRET",
		"DB_HOST",
		"DB_NAME",
		"DB_USER",
		"DB_PASSWORD",
	}

	missing := []string{}
	for _, key := range required {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		log.Fatalf("FATAL: Missing required environment variables for production: %v", missing)
	}

	// Warn about security-sensitive settings
	if os.Getenv("DB_SSLMODE") == "disable" {
		log.Println("WARNING: DB_SSLMODE is disabled - database connections are not encrypted")
	}
}
