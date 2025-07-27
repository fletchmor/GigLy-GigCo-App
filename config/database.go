package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	var err error

	// Load .env file if it exists (optional for Docker environments)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found in database config, using environment variables")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	log.Printf("Attempting to connect to database: host=%s port=%s dbname=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	// Retry connection with exponential backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		DB, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Test the connection
		err = DB.Ping()
		if err == nil {
			fmt.Println("Database connected successfully!")

			// Set connection pool settings
			DB.SetMaxOpenConns(25)
			DB.SetMaxIdleConns(25)
			DB.SetConnMaxLifetime(5 * time.Minute)

			return
		}

		log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	log.Fatal("Failed to connect to database after all retries:", err)
}
