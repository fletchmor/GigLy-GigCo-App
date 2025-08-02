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
	"strings"
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

// CreateJob handles job creation
func CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.JobCreateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateJobCreateRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Get consumer_id from JWT token
	// For now, we'll require it in the request or use a default
	consumerID := 2 // This should come from authentication

	// Insert job into database
	query := `
		INSERT INTO jobs (
			consumer_id, title, description, category, location_address,
			location_latitude, location_longitude, estimated_duration_hours,
			pay_rate_per_hour, total_pay, scheduled_start, scheduled_end, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, uuid, created_at, updated_at
	`

	var job model.Job
	err = config.DB.QueryRow(
		query,
		consumerID,
		req.Title,
		req.Description,
		nullString(req.Category),
		nullString(req.LocationAddress),
		nullFloat64Ptr(req.LocationLatitude),
		nullFloat64Ptr(req.LocationLongitude),
		nullFloat64Ptr(req.EstimatedDurationHours),
		nullFloat64Ptr(req.PayRatePerHour),
		nullFloat64Ptr(req.TotalPay),
		nullTimePtr(req.ScheduledStart),
		nullTimePtr(req.ScheduledEnd),
		nullString(req.Notes),
	).Scan(&job.ID, &job.UUID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		log.Printf("Database error creating job: %v", err)
		http.Error(w, "Failed to create job", http.StatusInternalServerError)
		return
	}

	// Populate the response with the request data
	job.ConsumerID = consumerID
	job.Title = req.Title
	job.Description = req.Description
	job.Category = req.Category
	job.LocationAddress = req.LocationAddress
	job.LocationLatitude = req.LocationLatitude
	job.LocationLongitude = req.LocationLongitude
	job.EstimatedDurationHours = req.EstimatedDurationHours
	job.PayRatePerHour = req.PayRatePerHour
	job.TotalPay = req.TotalPay
	job.ScheduledStart = req.ScheduledStart
	job.ScheduledEnd = req.ScheduledEnd
	job.Notes = req.Notes
	job.Status = "posted"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// GetJobs handles job listing with filters and pagination
func GetJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	status := r.URL.Query().Get("status")
	category := r.URL.Query().Get("category")
	consumerIDStr := r.URL.Query().Get("consumer_id")
	gigWorkerIDStr := r.URL.Query().Get("gig_worker_id")

	// Build dynamic query
	baseQuery := `
		SELECT j.id, j.uuid, j.consumer_id, j.gig_worker_id, j.title, j.description,
			   j.category, j.location_address, j.location_latitude, j.location_longitude,
			   j.estimated_duration_hours, j.pay_rate_per_hour, j.total_pay, j.status,
			   j.scheduled_start, j.scheduled_end, j.actual_start, j.actual_end,
			   j.notes, j.created_at, j.updated_at,
			   c.name as consumer_name, c.uuid as consumer_uuid
		FROM jobs j
		JOIN people c ON j.consumer_id = c.id
	`

	countQuery := "SELECT COUNT(*) FROM jobs j"

	var whereClauses []string
	var args []interface{}
	argIndex := 1

	// Add filters
	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("j.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("j.category = $%d", argIndex))
		args = append(args, category)
		argIndex++
	}

	if consumerIDStr != "" {
		if consumerID, err := strconv.Atoi(consumerIDStr); err == nil {
			whereClauses = append(whereClauses, fmt.Sprintf("j.consumer_id = $%d", argIndex))
			args = append(args, consumerID)
			argIndex++
		}
	}

	if gigWorkerIDStr != "" {
		if gigWorkerID, err := strconv.Atoi(gigWorkerIDStr); err == nil {
			whereClauses = append(whereClauses, fmt.Sprintf("j.gig_worker_id = $%d", argIndex))
			args = append(args, gigWorkerID)
			argIndex++
		}
	}

	// Add WHERE clause if we have filters
	if len(whereClauses) > 0 {
		whereClause := " WHERE " + strings.Join(whereClauses, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Printf("Error counting jobs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add pagination
	offset := (page - 1) * limit
	baseQuery += fmt.Sprintf(" ORDER BY j.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		log.Printf("Error querying jobs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jobs []model.JobResponse
	for rows.Next() {
		var job model.Job
		var consumerName, consumerUUID string

		err := rows.Scan(
			&job.ID, &job.UUID, &job.ConsumerID, &job.GigWorkerID, &job.Title, &job.Description,
			&job.Category, &job.LocationAddress, &job.LocationLatitude, &job.LocationLongitude,
			&job.EstimatedDurationHours, &job.PayRatePerHour, &job.TotalPay, &job.Status,
			&job.ScheduledStart, &job.ScheduledEnd, &job.ActualStart, &job.ActualEnd,
			&job.Notes, &job.CreatedAt, &job.UpdatedAt,
			&consumerName, &consumerUUID,
		)
		if err != nil {
			log.Printf("Error scanning job row: %v", err)
			continue
		}

		jobResponse := model.JobResponse{
			Job: job,
			Consumer: &model.UserSummary{
				ID:   job.ConsumerID,
				UUID: consumerUUID,
				Name: consumerName,
			},
		}

		jobs = append(jobs, jobResponse)
	}

	// Calculate pagination metadata
	pages := (total + limit - 1) / limit
	response := model.JobsListResponse{
		Jobs: jobs,
		Pagination: model.Pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			Pages:   pages,
			HasNext: page < pages,
			HasPrev: page > 1,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetJobByID retrieves a specific job by ID
func GetJobByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Invalid job ID format"})
		return
	}

	query := `
		SELECT j.id, j.uuid, j.consumer_id, j.gig_worker_id, j.title, j.description,
			   j.category, j.location_address, j.location_latitude, j.location_longitude,
			   j.estimated_duration_hours, j.pay_rate_per_hour, j.total_pay, j.status,
			   j.scheduled_start, j.scheduled_end, j.actual_start, j.actual_end,
			   j.notes, j.created_at, j.updated_at,
			   c.name as consumer_name, c.uuid as consumer_uuid,
			   w.name as worker_name, w.uuid as worker_uuid
		FROM jobs j
		JOIN people c ON j.consumer_id = c.id
		LEFT JOIN people w ON j.gig_worker_id = w.id
		WHERE j.id = $1
	`

	var job model.Job
	var consumerName, consumerUUID string
	var workerName, workerUUID sql.NullString

	err = config.DB.QueryRow(query, id).Scan(
		&job.ID, &job.UUID, &job.ConsumerID, &job.GigWorkerID, &job.Title, &job.Description,
		&job.Category, &job.LocationAddress, &job.LocationLatitude, &job.LocationLongitude,
		&job.EstimatedDurationHours, &job.PayRatePerHour, &job.TotalPay, &job.Status,
		&job.ScheduledStart, &job.ScheduledEnd, &job.ActualStart, &job.ActualEnd,
		&job.Notes, &job.CreatedAt, &job.UpdatedAt,
		&consumerName, &consumerUUID,
		&workerName, &workerUUID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Job not found"})
			return
		}
		log.Printf("Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Internal server error"})
		return
	}

	jobResponse := model.JobResponse{
		Job: job,
		Consumer: &model.UserSummary{
			ID:   job.ConsumerID,
			UUID: consumerUUID,
			Name: consumerName,
		},
	}

	// Add gig worker info if assigned
	if job.GigWorkerID != nil && workerName.Valid {
		jobResponse.GigWorker = &model.UserSummary{
			ID:   *job.GigWorkerID,
			UUID: workerUUID.String,
			Name: workerName.String,
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobResponse)
}

// AcceptJob allows a gig worker to accept a posted job
func AcceptJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idParam := chi.URLParam(r, "id")
	jobID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid job ID format", http.StatusBadRequest)
		return
	}

	// TODO: Get gig worker ID from JWT token
	// For now, we'll require it in the request body
	var req struct {
		GigWorkerID int `json:"gig_worker_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if req.GigWorkerID <= 0 {
		http.Error(w, "Gig worker ID is required", http.StatusBadRequest)
		return
	}

	// Update job with gig worker and change status
	query := `
		UPDATE jobs 
		SET gig_worker_id = $1, status = 'accepted', updated_at = NOW()
		WHERE id = $2 AND status = 'posted' AND gig_worker_id IS NULL
		RETURNING id, uuid, updated_at
	`

	var id int
	var uuid string
	var updatedAt time.Time

	err = config.DB.QueryRow(query, req.GigWorkerID, jobID).Scan(&id, &uuid, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found, already accepted, or not available", http.StatusConflict)
			return
		}
		log.Printf("Database error accepting job: %v", err)
		http.Error(w, "Failed to accept job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Job accepted successfully",
		"job_id":     id,
		"job_uuid":   uuid,
		"updated_at": updatedAt,
	})
}

// Helper functions for handling nullable database fields
func nullTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func nullFloat64Ptr(f *float64) interface{} {
	if f == nil {
		return nil
	}
	return *f
}

// validateJobCreateRequest validates the job creation request
func validateJobCreateRequest(req *model.JobCreateRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(req.Title) < 3 || len(req.Title) > 255 {
		return fmt.Errorf("title must be between 3 and 255 characters")
	}

	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(req.Description) < 10 {
		return fmt.Errorf("description must be at least 10 characters")
	}

	if req.EstimatedDurationHours != nil && *req.EstimatedDurationHours <= 0 {
		return fmt.Errorf("estimated duration must be greater than 0")
	}

	if req.PayRatePerHour != nil && *req.PayRatePerHour <= 0 {
		return fmt.Errorf("pay rate must be greater than 0")
	}

	if req.TotalPay != nil && *req.TotalPay <= 0 {
		return fmt.Errorf("total pay must be greater than 0")
	}

	// Validate time constraints
	if req.ScheduledStart != nil && req.ScheduledEnd != nil {
		if req.ScheduledEnd.Before(*req.ScheduledStart) {
			return fmt.Errorf("scheduled end time must be after start time")
		}
	}

	return nil
}
