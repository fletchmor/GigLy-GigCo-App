package api

import (
	"app/config"
	"app/internal/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Test database connection
	err := config.DB.Ping()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"database":  "connected",
		"timestamp": time.Now(),
	})
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if user.Name == "" || user.Address == "" {
		http.Error(w, "Name and address are required", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO customers (name, address, created_at) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at`

	err = config.DB.QueryRow(query, user.Name, user.Address, time.Now()).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var customer model.User
	query := "SELECT id, name FROM customers WHERE id = $1"
	err = config.DB.QueryRow(query, id).Scan(&customer.ID, &customer.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Customer not found"})
			return
		}
		log.Printf("Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(customer)
}

func CreateSchedule(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the request body into a Schedule struct
	var schedule model.Schedule
	err := json.NewDecoder(r.Body).Decode(&schedule)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if schedule.GigWorkerID <= 0 {
		http.Error(w, "Gig worker ID is required", http.StatusBadRequest)
		return
	}
	if schedule.StartTime.IsZero() {
		http.Error(w, "Start time is required", http.StatusBadRequest)
		return
	}
	if schedule.EndTime.IsZero() {
		http.Error(w, "End time is required", http.StatusBadRequest)
		return
	}
	if schedule.StartTime.After(schedule.EndTime) {
		http.Error(w, "Start time must be before end time", http.StatusBadRequest)
		return
	}

	// Check if job_id is provided and exists in the jobs table
	if schedule.JobID != nil {
		var exists bool
		err := config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM jobs WHERE id = $1)", *schedule.JobID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking job existence: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			log.Printf("Invalid job_id: %d does not exist", *schedule.JobID)
			http.Error(w, "Invalid job_id: the specified job does not exist", http.StatusBadRequest)
			return
		}
	}

	// Insert the schedule into the database
	query := `
        INSERT INTO schedules (
            gig_worker_id, title, start_time, end_time, is_available, job_id,
            recurring_pattern, recurring_until, notes
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9
        ) RETURNING id, uuid, created_at, updated_at
    `

	var id int
	var uuid string
	var createdAt time.Time
	var updatedAt time.Time

	err = config.DB.QueryRow(query,
		schedule.GigWorkerID,
		schedule.Title,
		schedule.StartTime,
		schedule.EndTime,
		schedule.IsAvailable,
		schedule.JobID, // Will be NULL if schedule.JobID is nil
		schedule.RecurringPattern,
		schedule.RecurringUntil,
		schedule.Notes,
	).Scan(&id, &uuid, &createdAt, &updatedAt)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Failed to create schedule", http.StatusInternalServerError)
		return
	}

	// Populate the generated fields in the response
	schedule.ID = id
	schedule.Uuid = uuid
	schedule.CreatedAt = createdAt
	schedule.UpdatedAt = updatedAt

	// Return the created schedule as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var transaction model.Transaction
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if transaction.JobID <= 0 {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}
	if transaction.ConsumerID <= 0 {
		http.Error(w, "Consumer ID is required", http.StatusBadRequest)
		return
	}
	if transaction.GigWorkerID <= 0 {
		http.Error(w, "Gig worker ID is required", http.StatusBadRequest)
		return
	}
	if transaction.Amount <= 0 {
		http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
		return
	}
	if transaction.PaymentMethod == "" {
		http.Error(w, "Payment method is required", http.StatusBadRequest)
		return
	}

	// Set default values
	if transaction.Currency == "" {
		transaction.Currency = "USD"
	}
	if transaction.Status == "" {
		transaction.Status = "pending"
	}

	// Insert into database
	query := `
        INSERT INTO transactions (
            job_id, consumer_id, gig_worker_id, amount, currency, status,
            payment_method, notes
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8
        ) RETURNING id, uuid, created_at, updated_at
    `

	var id int
	var uuid string
	var createdAt time.Time
	var updatedAt time.Time

	err = config.DB.QueryRow(query,
		transaction.JobID,
		transaction.ConsumerID,
		transaction.GigWorkerID,
		transaction.Amount,
		transaction.Currency,
		transaction.Status,
		transaction.PaymentMethod,
		transaction.Notes,
	).Scan(&id, &uuid, &createdAt, &updatedAt)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	// Set generated fields
	transaction.ID = id
	transaction.Uuid = uuid
	transaction.CreatedAt = createdAt
	transaction.UpdatedAt = updatedAt

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}
