package activities

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	"app/internal/temporal/workflows"
)

// JobActivities contains all job-related activities
type JobActivities struct {
	db *sql.DB
}

// NewJobActivities creates a new JobActivities instance
func NewJobActivities(db *sql.DB) *JobActivities {
	return &JobActivities{db: db}
}

// PriceJob calculates the price for a job based on requirements
func (a *JobActivities) PriceJob(ctx context.Context, jobID int) (workflows.PriceJobResult, error) {
	log.Printf("Pricing job %d", jobID)

	// Get job details from database
	var job struct {
		ID          int
		Title       string
		Description string
		Duration    int // in hours
		Skills      string
		Urgency     string
		Location    string
	}

	query := `
		SELECT id, title, description, 
		       COALESCE(estimated_duration_hours, 1) as duration,
		       COALESCE(category, '') as skills,
		       'medium' as urgency,
		       COALESCE(location_address, '') as location
		FROM jobs WHERE id = $1
	`
	err := a.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.Title, &job.Description, &job.Duration,
		&job.Skills, &job.Urgency, &job.Location,
	)
	if err != nil {
		return workflows.PriceJobResult{}, fmt.Errorf("failed to get job details: %w", err)
	}

	// Calculate base price
	baseRate := 25.0 // $25/hour base rate
	totalPrice := baseRate * float64(job.Duration)

	// Apply urgency multiplier
	switch job.Urgency {
	case "urgent":
		totalPrice *= 1.5
	case "high":
		totalPrice *= 1.3
	case "medium":
		totalPrice *= 1.1
	}

	// Round to nearest dollar
	totalPrice = math.Round(totalPrice*100) / 100

	// Update job with calculated price
	updateQuery := `
		UPDATE jobs 
		SET total_pay = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`
	_, err = a.db.ExecContext(ctx, updateQuery, totalPrice, jobID)
	if err != nil {
		return workflows.PriceJobResult{}, fmt.Errorf("failed to update job price: %w", err)
	}

	log.Printf("Job %d priced at $%.2f", jobID, totalPrice)

	return workflows.PriceJobResult{
		JobID:  jobID,
		Amount: totalPrice,
	}, nil
}

// SendJobOffer sends a job offer to the customer
func (a *JobActivities) SendJobOffer(ctx context.Context, jobID int, amount float64) error {
	log.Printf("Sending job offer for job %d with amount $%.2f", jobID, amount)

	// Update job status to indicate offer sent
	query := `
		UPDATE jobs 
		SET status = 'offer_sent', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err := a.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// In a real implementation, you would:
	// 1. Send email/SMS to customer
	// 2. Create notification in app
	// 3. Log the offer in audit table

	log.Printf("Job offer sent successfully for job %d", jobID)
	return nil
}

// FindMatchingWorker finds an available worker for the job
func (a *JobActivities) FindMatchingWorker(ctx context.Context, jobID int) (workflows.MatchWorkerResult, error) {
	log.Printf("Finding matching worker for job %d", jobID)

	// Get job requirements
	var jobSkills, jobLocation string
	err := a.db.QueryRowContext(ctx,
		"SELECT COALESCE(category, '') as skills, COALESCE(location_address, '') as location FROM jobs WHERE id = $1",
		jobID).Scan(&jobSkills, &jobLocation)
	if err != nil {
		return workflows.MatchWorkerResult{}, fmt.Errorf("failed to get job details: %w", err)
	}

	// Find available workers
	// This is a simplified matching algorithm
	query := `
		SELECT gw.id, gw.name, COALESCE(gw.bio, '') as skills, 
		       COALESCE(gw.address, '') as location, 5.0 as rating
		FROM gigworkers gw
		WHERE gw.is_active = true
		ORDER BY gw.created_at ASC
		LIMIT 5
	`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return workflows.MatchWorkerResult{}, fmt.Errorf("failed to query workers: %w", err)
	}
	defer rows.Close()

	var bestWorkerID int
	var bestRating float64

	for rows.Next() {
		var workerID int
		var name, skills, location string
		var rating float64

		err := rows.Scan(&workerID, &name, &skills, &location, &rating)
		if err != nil {
			log.Printf("Error scanning worker row: %v", err)
			continue
		}

		// Simple matching: take the highest rated available worker
		if rating > bestRating {
			bestWorkerID = workerID
			bestRating = rating
		}
	}

	if bestWorkerID == 0 {
		return workflows.MatchWorkerResult{}, fmt.Errorf("no available workers found")
	}

	// Assign worker to job
	updateQuery := `
		UPDATE jobs 
		SET gig_worker_id = $1, status = 'worker_assigned', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`
	_, err = a.db.ExecContext(ctx, updateQuery, bestWorkerID, jobID)
	if err != nil {
		return workflows.MatchWorkerResult{}, fmt.Errorf("failed to assign worker: %w", err)
	}

	// Mark worker as unavailable
	_, err = a.db.ExecContext(ctx,
		"UPDATE gigworkers SET is_active = false WHERE id = $1",
		bestWorkerID)
	if err != nil {
		log.Printf("Warning: failed to mark worker as unavailable: %v", err)
	}

	log.Printf("Worker %d assigned to job %d", bestWorkerID, jobID)

	return workflows.MatchWorkerResult{
		JobID:    jobID,
		WorkerID: bestWorkerID,
	}, nil
}

// ScheduleJob schedules the job with the assigned worker
func (a *JobActivities) ScheduleJob(ctx context.Context, jobID, workerID int) error {
	log.Printf("Scheduling job %d with worker %d", jobID, workerID)

	// Create a schedule entry
	// For now, schedule for tomorrow at 9 AM
	scheduledTime := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(9 * time.Hour)

	query := `
		INSERT INTO schedules (gig_worker_id, title, start_time, end_time, is_available, job_id, created_at)
		VALUES ($1, $2, $3, $4, false, $5, CURRENT_TIMESTAMP)
	`
	endTime := scheduledTime.Add(2 * time.Hour) // 2 hour job duration
	_, err := a.db.ExecContext(ctx, query, workerID, "Scheduled Job", scheduledTime, endTime, jobID)
	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	// Update job status
	updateQuery := `
		UPDATE jobs 
		SET status = 'scheduled', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err = a.db.ExecContext(ctx, updateQuery, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	log.Printf("Job %d scheduled for %v", jobID, scheduledTime)
	return nil
}

// ProcessJobPayment processes payment for a completed job
func (a *JobActivities) ProcessJobPayment(ctx context.Context, jobID int) (workflows.ProcessPaymentResult, error) {
	log.Printf("Processing payment for job %d", jobID)

	// Get job and payment details
	var job struct {
		ID         int
		ConsumerID int
		WorkerID   int
		TotalPay   float64
		Status     string
	}

	query := `
		SELECT id, consumer_id, gig_worker_id, total_pay, status
		FROM jobs WHERE id = $1
	`
	err := a.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.ConsumerID, &job.WorkerID, &job.TotalPay, &job.Status,
	)
	if err != nil {
		return workflows.ProcessPaymentResult{}, fmt.Errorf("failed to get job details: %w", err)
	}

	if job.Status != "completed" {
		return workflows.ProcessPaymentResult{}, fmt.Errorf("job not completed, cannot process payment")
	}

	// Create transaction record
	transactionID := fmt.Sprintf("txn_%d_%d", jobID, time.Now().Unix())

	insertQuery := `
		INSERT INTO transactions (job_id, consumer_id, gig_worker_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, 'completed', CURRENT_TIMESTAMP)
	`
	_, err = a.db.ExecContext(ctx, insertQuery,
		job.ID, job.ConsumerID, job.WorkerID, job.TotalPay)
	if err != nil {
		return workflows.ProcessPaymentResult{}, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update job status
	updateQuery := `
		UPDATE jobs 
		SET status = 'paid', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err = a.db.ExecContext(ctx, updateQuery, jobID)
	if err != nil {
		return workflows.ProcessPaymentResult{}, fmt.Errorf("failed to update job status: %w", err)
	}

	// Mark worker as available again
	_, err = a.db.ExecContext(ctx,
		"UPDATE gigworkers SET is_active = true WHERE id = $1",
		job.WorkerID)
	if err != nil {
		log.Printf("Warning: failed to mark worker as available: %v", err)
	}

	log.Printf("Payment processed for job %d, transaction %s", jobID, transactionID)

	return workflows.ProcessPaymentResult{
		TransactionID: transactionID,
		Amount:        job.TotalPay,
	}, nil
}

// RequestReviews sends review requests to both consumer and worker
func (a *JobActivities) RequestReviews(ctx context.Context, jobID int) error {
	log.Printf("Requesting reviews for job %d", jobID)

	// Update job status
	query := `
		UPDATE jobs 
		SET status = 'review_pending', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err := a.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// In a real implementation, you would:
	// 1. Send review request emails/notifications
	// 2. Create review reminder tasks
	// 3. Set up review deadline tracking

	log.Printf("Review requests sent for job %d", jobID)
	return nil
}

// CloseJob finalizes the job
func (a *JobActivities) CloseJob(ctx context.Context, jobID int) error {
	log.Printf("Closing job %d", jobID)

	query := `
		UPDATE jobs
		SET status = 'closed', workflow_completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := a.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to close job: %w", err)
	}

	log.Printf("Job %d closed successfully", jobID)
	return nil
}

// HandleJobRejection handles when a customer rejects a job offer
func (a *JobActivities) HandleJobRejection(ctx context.Context, jobID int) error {
	log.Printf("Handling job rejection for job %d", jobID)

	query := `
		UPDATE jobs 
		SET status = 'rejected', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err := a.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	log.Printf("Job %d marked as rejected", jobID)
	return nil
}

// HandleNoWorkerAvailable handles when no worker is available
func (a *JobActivities) HandleNoWorkerAvailable(ctx context.Context, jobID int) error {
	log.Printf("Handling no worker available for job %d", jobID)

	query := `
		UPDATE jobs 
		SET status = 'no_worker_available', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err := a.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	log.Printf("Job %d marked as no worker available", jobID)
	return nil
}

// HandlePaymentFailure handles payment processing failures
func (a *JobActivities) HandlePaymentFailure(ctx context.Context, jobID int) error {
	log.Printf("Handling payment failure for job %d", jobID)

	query := `
		UPDATE jobs 
		SET status = 'payment_failed', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err := a.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	log.Printf("Job %d marked as payment failed", jobID)
	return nil
}

// UpdateJobPaymentStatus updates job payment status after successful retry
func (a *JobActivities) UpdateJobPaymentStatus(ctx context.Context, jobID int, transactionID string) error {
	log.Printf("Updating payment status for job %d with transaction %s", jobID, transactionID)

	query := `
		UPDATE jobs 
		SET status = 'paid', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`
	_, err := a.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job payment status: %w", err)
	}

	log.Printf("Job %d payment status updated", jobID)
	return nil
}
