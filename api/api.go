package api

import (
	"app/config"
	"app/internal/model"
	"app/internal/temporal"
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
	query := "SELECT id, name, COALESCE(address, '') as address FROM customers WHERE id = $1"
	err = config.DB.QueryRow(query, id).Scan(&customer.ID, &customer.Name, &customer.Address)
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

	// Validate foreign key relationships exist
	var exists bool
	
	// Check if job exists
	err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM jobs WHERE id = $1)", transaction.JobID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking job existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Job not found", http.StatusBadRequest)
		return
	}
	
	// Check if consumer exists
	err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM people WHERE id = $1)", transaction.ConsumerID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking consumer existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Consumer not found", http.StatusBadRequest)
		return
	}
	
	// Check if gig worker exists
	err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM people WHERE id = $1)", transaction.GigWorkerID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking gig worker existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Gig worker not found", http.StatusBadRequest)
		return
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
		log.Printf("Database error creating transaction: %v", err)
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
	// For now, use from request (required for API tests)
	consumerID := req.ConsumerID
	if consumerID == 0 {
		http.Error(w, "Consumer ID is required", http.StatusBadRequest)
		return
	}

	// Handle alternative field names for backward compatibility
	locationAddress := req.LocationAddress
	if locationAddress == "" && req.Location != "" {
		locationAddress = req.Location
	}
	
	var estimatedHours *float64
	if req.EstimatedDurationHours != nil {
		estimatedHours = req.EstimatedDurationHours
	} else if req.EstimatedHours != nil {
		estimatedHours = req.EstimatedHours
	}
	
	var payRate *float64
	if req.PayRatePerHour != nil {
		payRate = req.PayRatePerHour
	} else if req.PayRate != nil {
		payRate = req.PayRate
	}

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
		nullString(locationAddress),
		nullFloat64Ptr(req.LocationLatitude),
		nullFloat64Ptr(req.LocationLongitude),
		nullFloat64Ptr(estimatedHours),
		nullFloat64Ptr(payRate),
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

	// Populate the response with the processed data
	job.ConsumerID = consumerID
	job.Title = req.Title
	job.Description = req.Description
	job.Category = req.Category
	job.LocationAddress = locationAddress
	job.LocationLatitude = req.LocationLatitude
	job.LocationLongitude = req.LocationLongitude
	job.EstimatedDurationHours = estimatedHours
	job.PayRatePerHour = payRate
	job.TotalPay = req.TotalPay
	job.ScheduledStart = req.ScheduledStart
	job.ScheduledEnd = req.ScheduledEnd
	job.Notes = req.Notes
	job.Status = "posted"

	// Start Temporal workflow for the job asynchronously to avoid blocking the response
	go func() {
		temporalClient, err := temporal.NewClient()
		if err != nil {
			log.Printf("Failed to create Temporal client: %v", err)
			return
		}
		defer temporalClient.Close()
		
		we, err := temporalClient.StartJobWorkflow(r.Context(), job.ID, job.ConsumerID)
		if err != nil {
			log.Printf("Failed to start job workflow: %v", err)
			return
		}
		
		// Update job with workflow information
		updateQuery := `
			UPDATE jobs 
			SET temporal_workflow_id = $1, temporal_run_id = $2, updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`
		_, err = config.DB.Exec(updateQuery, we.GetID(), we.GetRunID(), job.ID)
		if err != nil {
			log.Printf("Failed to update job with workflow IDs: %v", err)
		} else {
			log.Printf("Started workflow for job %d: %s", job.ID, we.GetID())
		}
	}()

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

	// Check if job exists first
	var existingStatus sql.NullString
	var existingGigWorkerID sql.NullInt32
	checkQuery := "SELECT status, gig_worker_id FROM jobs WHERE id = $1"
	err = config.DB.QueryRow(checkQuery, jobID).Scan(&existingStatus, &existingGigWorkerID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error checking job: %v", err)
		http.Error(w, "Failed to check job status", http.StatusInternalServerError)
		return
	}

	// Check if job is in correct status and available
	if !existingStatus.Valid || existingStatus.String != "posted" {
		http.Error(w, "Job is not available for acceptance", http.StatusConflict)
		return
	}
	
	if existingGigWorkerID.Valid {
		http.Error(w, "Job has already been accepted by another worker", http.StatusConflict)
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
			http.Error(w, "Job acceptance failed due to concurrent update", http.StatusConflict)
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

// CreateGigWorker handles gig worker creation
func CreateGigWorker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var gigWorker model.GigWorker
	err := json.NewDecoder(r.Body).Decode(&gigWorker)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateGigWorkerRequest(&gigWorker); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set defaults
	gigWorker.Role = "gig_worker"
	// Default to active (JSON decoder sets bool to false, so we override)
	gigWorker.IsActive = true
	if gigWorker.VerificationStatus == "" {
		gigWorker.VerificationStatus = "pending"
	}

	// Insert into gigworkers table
	query := `
		INSERT INTO gigworkers (
			name, email, phone, address, latitude, longitude, place_id, 
			role, is_active, email_verified, phone_verified, bio, hourly_rate, 
			experience_years, verification_status, background_check_date, 
			service_radius_miles, availability_notes, emergency_contact_name, 
			emergency_contact_phone, emergency_contact_relationship, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		) RETURNING id, uuid, created_at, updated_at`

	var id int
	var uuid string
	var createdAt, updatedAt time.Time
	now := time.Now()

	err = config.DB.QueryRow(
		query,
		gigWorker.Name,
		gigWorker.Email,
		nullString(gigWorker.Phone),
		gigWorker.Address,
		nullFloat64(gigWorker.Latitude),
		nullFloat64(gigWorker.Longitude),
		nullString(gigWorker.PlaceID),
		gigWorker.Role,
		gigWorker.IsActive,
		gigWorker.EmailVerified,
		gigWorker.PhoneVerified,
		nullString(gigWorker.Bio),
		nullFloat64Ptr(gigWorker.HourlyRate),
		nullIntPtr(gigWorker.ExperienceYears),
		gigWorker.VerificationStatus,
		nullTimePtr(gigWorker.BackgroundCheckDate),
		nullFloat64Ptr(gigWorker.ServiceRadiusMiles),
		nullString(gigWorker.AvailabilityNotes),
		nullString(gigWorker.EmergencyContactName),
		nullString(gigWorker.EmergencyContactPhone),
		nullString(gigWorker.EmergencyContactRelationship),
		now,
		now,
	).Scan(&id, &uuid, &createdAt, &updatedAt)

	if err != nil {
		log.Printf("Database error creating gig worker: %v", err)
		http.Error(w, "Failed to create gig worker", http.StatusInternalServerError)
		return
	}

	// Set response fields
	gigWorker.ID = id
	gigWorker.Uuid = uuid
	gigWorker.CreatedAt = createdAt
	gigWorker.UpdatedAt = updatedAt

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gigWorker)
}

// validateGigWorkerRequest validates the gig worker creation request
func validateGigWorkerRequest(gw *model.GigWorker) error {
	if gw.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(gw.Name) < 2 || len(gw.Name) > 255 {
		return fmt.Errorf("name must be between 2 and 255 characters")
	}

	if gw.Email == "" {
		return fmt.Errorf("email is required")
	}
	if len(gw.Email) > 255 {
		return fmt.Errorf("email must be less than 255 characters")
	}

	if gw.Address == "" {
		return fmt.Errorf("address is required")
	}

	if gw.HourlyRate != nil && *gw.HourlyRate <= 0 {
		return fmt.Errorf("hourly rate must be greater than 0")
	}

	if gw.ExperienceYears != nil && (*gw.ExperienceYears < 0 || *gw.ExperienceYears > 50) {
		return fmt.Errorf("experience years must be between 0 and 50")
	}

	if gw.ServiceRadiusMiles != nil && (*gw.ServiceRadiusMiles < 1 || *gw.ServiceRadiusMiles > 100) {
		return fmt.Errorf("service radius must be between 1 and 100 miles")
	}

	return nil
}

// GetGigWorkers handles retrieving all gig workers with optional filtering
func GetGigWorkers(w http.ResponseWriter, r *http.Request) {
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

	verificationStatus := r.URL.Query().Get("verification_status")
	isActive := r.URL.Query().Get("is_active")

	// Build dynamic query
	baseQuery := `
		SELECT id, uuid, name, email, phone, address, latitude, longitude, place_id,
			   role, is_active, email_verified, phone_verified, bio, hourly_rate,
			   experience_years, verification_status, background_check_date,
			   service_radius_miles, availability_notes, emergency_contact_name,
			   emergency_contact_phone, emergency_contact_relationship, created_at, updated_at
		FROM gigworkers
	`

	countQuery := "SELECT COUNT(*) FROM gigworkers"

	var whereClauses []string
	var args []interface{}
	argIndex := 1

	// Add filters
	if verificationStatus != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("verification_status = $%d", argIndex))
		args = append(args, verificationStatus)
		argIndex++
	}

	if isActive != "" {
		if isActive == "true" {
			whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, true)
		} else if isActive == "false" {
			whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, false)
		}
		argIndex++
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
		log.Printf("Error counting gig workers: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add pagination
	offset := (page - 1) * limit
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		log.Printf("Error querying gig workers: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var gigWorkers []model.GigWorker
	for rows.Next() {
		var gw model.GigWorker
		var phone, placeID, bio, availabilityNotes sql.NullString
		var latitude, longitude sql.NullFloat64
		var hourlyRate, serviceRadiusMiles sql.NullFloat64
		var experienceYears sql.NullInt32
		var backgroundCheckDate sql.NullTime
		var emergencyContactName, emergencyContactPhone, emergencyContactRelationship sql.NullString

		err := rows.Scan(
			&gw.ID, &gw.Uuid, &gw.Name, &gw.Email, &phone, &gw.Address,
			&latitude, &longitude, &placeID, &gw.Role, &gw.IsActive,
			&gw.EmailVerified, &gw.PhoneVerified, &bio, &hourlyRate,
			&experienceYears, &gw.VerificationStatus, &backgroundCheckDate,
			&serviceRadiusMiles, &availabilityNotes, &emergencyContactName,
			&emergencyContactPhone, &emergencyContactRelationship,
			&gw.CreatedAt, &gw.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning gig worker row: %v", err)
			continue
		}

		// Handle nullable fields
		if phone.Valid {
			gw.Phone = phone.String
		}
		if placeID.Valid {
			gw.PlaceID = placeID.String
		}
		if latitude.Valid {
			gw.Latitude = latitude.Float64
		}
		if longitude.Valid {
			gw.Longitude = longitude.Float64
		}
		if bio.Valid {
			gw.Bio = bio.String
		}
		if hourlyRate.Valid {
			gw.HourlyRate = &hourlyRate.Float64
		}
		if experienceYears.Valid {
			years := int(experienceYears.Int32)
			gw.ExperienceYears = &years
		}
		if backgroundCheckDate.Valid {
			gw.BackgroundCheckDate = &backgroundCheckDate.Time
		}
		if serviceRadiusMiles.Valid {
			gw.ServiceRadiusMiles = &serviceRadiusMiles.Float64
		}
		if availabilityNotes.Valid {
			gw.AvailabilityNotes = availabilityNotes.String
		}
		if emergencyContactName.Valid {
			gw.EmergencyContactName = emergencyContactName.String
		}
		if emergencyContactPhone.Valid {
			gw.EmergencyContactPhone = emergencyContactPhone.String
		}
		if emergencyContactRelationship.Valid {
			gw.EmergencyContactRelationship = emergencyContactRelationship.String
		}

		gigWorkers = append(gigWorkers, gw)
	}

	// Calculate pagination metadata
	pages := (total + limit - 1) / limit
	response := map[string]interface{}{
		"gigworkers": gigWorkers,
		"pagination": model.Pagination{
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

// GetGigWorkerByID retrieves a specific gig worker by ID
func GetGigWorkerByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Invalid gig worker ID format"})
		return
	}

	query := `
		SELECT id, uuid, name, email, phone, address, latitude, longitude, place_id,
			   role, is_active, email_verified, phone_verified, bio, hourly_rate,
			   experience_years, verification_status, background_check_date,
			   service_radius_miles, availability_notes, emergency_contact_name,
			   emergency_contact_phone, emergency_contact_relationship, created_at, updated_at
		FROM gigworkers
		WHERE id = $1
	`

	var gw model.GigWorker
	var phone, placeID, bio, availabilityNotes sql.NullString
	var latitude, longitude sql.NullFloat64
	var hourlyRate, serviceRadiusMiles sql.NullFloat64
	var experienceYears sql.NullInt32
	var backgroundCheckDate sql.NullTime
	var emergencyContactName, emergencyContactPhone, emergencyContactRelationship sql.NullString

	err = config.DB.QueryRow(query, id).Scan(
		&gw.ID, &gw.Uuid, &gw.Name, &gw.Email, &phone, &gw.Address,
		&latitude, &longitude, &placeID, &gw.Role, &gw.IsActive,
		&gw.EmailVerified, &gw.PhoneVerified, &bio, &hourlyRate,
		&experienceYears, &gw.VerificationStatus, &backgroundCheckDate,
		&serviceRadiusMiles, &availabilityNotes, &emergencyContactName,
		&emergencyContactPhone, &emergencyContactRelationship,
		&gw.CreatedAt, &gw.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Gig worker not found"})
			return
		}
		log.Printf("Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Internal server error"})
		return
	}

	// Handle nullable fields
	if phone.Valid {
		gw.Phone = phone.String
	}
	if placeID.Valid {
		gw.PlaceID = placeID.String
	}
	if latitude.Valid {
		gw.Latitude = latitude.Float64
	}
	if longitude.Valid {
		gw.Longitude = longitude.Float64
	}
	if bio.Valid {
		gw.Bio = bio.String
	}
	if hourlyRate.Valid {
		gw.HourlyRate = &hourlyRate.Float64
	}
	if experienceYears.Valid {
		years := int(experienceYears.Int32)
		gw.ExperienceYears = &years
	}
	if backgroundCheckDate.Valid {
		gw.BackgroundCheckDate = &backgroundCheckDate.Time
	}
	if serviceRadiusMiles.Valid {
		gw.ServiceRadiusMiles = &serviceRadiusMiles.Float64
	}
	if availabilityNotes.Valid {
		gw.AvailabilityNotes = availabilityNotes.String
	}
	if emergencyContactName.Valid {
		gw.EmergencyContactName = emergencyContactName.String
	}
	if emergencyContactPhone.Valid {
		gw.EmergencyContactPhone = emergencyContactPhone.String
	}
	if emergencyContactRelationship.Valid {
		gw.EmergencyContactRelationship = emergencyContactRelationship.String
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gw)
}

// Helper functions for handling nullable database fields
func nullIntPtr(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
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
