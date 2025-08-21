package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"app/internal/temporal/activities"
	"app/internal/temporal/workflows"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting Temporal worker...")

	// Get database connection
	db, err := connectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Successfully connected to database")

	// Create Temporal client
	temporalHost := getEnv("TEMPORAL_HOST", "localhost:7233")
	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		log.Fatal("Unable to create Temporal client:", err)
	}
	defer c.Close()

	log.Printf("Connected to Temporal server at %s", temporalHost)

	// Create worker
	taskQueue := "gigco-jobs"
	w := worker.New(c, taskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.JobLifecycleWorkflow)
	w.RegisterWorkflow(workflows.PaymentRetryWorkflow)

	// Register activities
	jobActivities := activities.NewJobActivities(db)
	w.RegisterActivity(jobActivities.PriceJob)
	w.RegisterActivity(jobActivities.SendJobOffer)
	w.RegisterActivity(jobActivities.FindMatchingWorker)
	w.RegisterActivity(jobActivities.ScheduleJob)
	w.RegisterActivity(jobActivities.ProcessJobPayment)
	w.RegisterActivity(jobActivities.RequestReviews)
	w.RegisterActivity(jobActivities.CloseJob)
	w.RegisterActivity(jobActivities.HandleJobRejection)
	w.RegisterActivity(jobActivities.HandleNoWorkerAvailable)
	w.RegisterActivity(jobActivities.HandlePaymentFailure)
	w.RegisterActivity(jobActivities.UpdateJobPaymentStatus)

	log.Printf("Worker registered for task queue: %s", taskQueue)
	log.Println("Registered workflows: JobLifecycleWorkflow, PaymentRetryWorkflow")
	log.Println("Registered activities: PriceJob, SendJobOffer, FindMatchingWorker, ScheduleJob, ProcessJobPayment, RequestReviews, CloseJob, HandleJobRejection, HandleNoWorkerAvailable, HandlePaymentFailure, UpdateJobPaymentStatus")

	// Start worker
	log.Println("Starting worker...")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatal("Unable to start worker:", err)
	}

	log.Println("Worker stopped")
}

// connectDB creates a database connection using environment variables
func connectDB() (*sql.DB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "bamboo")
	dbName := getEnv("DB_NAME", "gigco")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	log.Printf("Connecting to database: host=%s port=%s dbname=%s user=%s sslmode=%s",
		dbHost, dbPort, dbName, dbUser, dbSSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
