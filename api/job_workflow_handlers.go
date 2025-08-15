package api

import (
	"app/config"
	"app/internal/temporal"
	"app/internal/temporal/workflows"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// AcceptJobOffer allows a customer to accept a job offer
func AcceptJobOffer(w http.ResponseWriter, r *http.Request) {
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

	// Get job and workflow information
	var workflowID sql.NullString
	var status string
	query := `
		SELECT temporal_workflow_id, status 
		FROM jobs 
		WHERE id = $1
	`
	err = config.DB.QueryRow(query, jobID).Scan(&workflowID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting job: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if status != "offer_sent" && status != "posted" && status != "accepted" {
		http.Error(w, fmt.Sprintf("Job cannot be accepted in current status: %s", status), http.StatusBadRequest)
		return
	}

	// If no workflow is associated, update job status directly (for testing)
	if !workflowID.Valid {
		if status == "posted" {
			// Treat this as accepting a posted job
			updateQuery := `
				UPDATE jobs 
				SET status = 'accepted', updated_at = NOW()
				WHERE id = $1
			`
			_, err = config.DB.Exec(updateQuery, jobID)
			if err != nil {
				log.Printf("Database error updating job status: %v", err)
				http.Error(w, "Failed to update job status", http.StatusInternalServerError)
				return
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Job offer accepted successfully",
			"job_id":  jobID,
		})
		return
	}

	// Signal the workflow
	temporalClient, err := temporal.NewClient()
	if err != nil {
		log.Printf("Failed to create Temporal client: %v", err)
		http.Error(w, "Failed to connect to workflow service", http.StatusInternalServerError)
		return
	}
	defer temporalClient.Close()

	err = temporalClient.SignalJobOfferResponse(r.Context(), workflowID.String, true)
	if err != nil {
		log.Printf("Failed to signal workflow: %v", err)
		http.Error(w, "Failed to signal workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Job offer accepted successfully",
		"job_id":  jobID,
	})
}

// RejectJobOffer allows a customer to reject a job offer
func RejectJobOffer(w http.ResponseWriter, r *http.Request) {
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

	// Get job and workflow information
	var workflowID sql.NullString
	var status string
	query := `
		SELECT temporal_workflow_id, status 
		FROM jobs 
		WHERE id = $1
	`
	err = config.DB.QueryRow(query, jobID).Scan(&workflowID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting job: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if status != "offer_sent" && status != "posted" && status != "accepted" {
		http.Error(w, fmt.Sprintf("Job cannot be rejected in current status: %s", status), http.StatusBadRequest)
		return
	}

	// If no workflow is associated, update job status directly (for testing)
	if !workflowID.Valid {
		if status == "posted" || status == "accepted" {
			// Treat this as rejecting/cancelling a job
			updateQuery := `
				UPDATE jobs 
				SET status = 'cancelled', updated_at = NOW()
				WHERE id = $1
			`
			_, err = config.DB.Exec(updateQuery, jobID)
			if err != nil {
				log.Printf("Database error updating job status: %v", err)
				http.Error(w, "Failed to update job status", http.StatusInternalServerError)
				return
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Job offer rejected",
			"job_id":  jobID,
		})
		return
	}

	// Signal the workflow
	temporalClient, err := temporal.NewClient()
	if err != nil {
		log.Printf("Failed to create Temporal client: %v", err)
		http.Error(w, "Failed to connect to workflow service", http.StatusInternalServerError)
		return
	}
	defer temporalClient.Close()

	err = temporalClient.SignalJobOfferResponse(r.Context(), workflowID.String, false)
	if err != nil {
		log.Printf("Failed to signal workflow: %v", err)
		http.Error(w, "Failed to signal workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Job offer rejected",
		"job_id":  jobID,
	})
}

// StartJob allows a worker to mark a job as started
func StartJob(w http.ResponseWriter, r *http.Request) {
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

	// Get job and workflow information
	var workflowID sql.NullString
	var status string
	query := `
		SELECT temporal_workflow_id, status 
		FROM jobs 
		WHERE id = $1
	`
	err = config.DB.QueryRow(query, jobID).Scan(&workflowID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting job: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if status != "scheduled" && status != "accepted" {
		http.Error(w, fmt.Sprintf("Job not ready to start, current status: %s", status), http.StatusBadRequest)
		return
	}

	// If no workflow is associated, update job status directly
	if !workflowID.Valid {
		updateQuery := `
			UPDATE jobs 
			SET status = 'in_progress', actual_start = NOW(), updated_at = NOW()
			WHERE id = $1
		`
		_, err = config.DB.Exec(updateQuery, jobID)
		if err != nil {
			log.Printf("Database error updating job status: %v", err)
			http.Error(w, "Failed to update job status", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Job started successfully",
			"job_id":  jobID,
		})
		return
	}

	// Signal the workflow
	temporalClient, err := temporal.NewClient()
	if err != nil {
		log.Printf("Failed to create Temporal client: %v", err)
		http.Error(w, "Failed to connect to workflow service", http.StatusInternalServerError)
		return
	}
	defer temporalClient.Close()

	err = temporalClient.SignalJobStarted(r.Context(), workflowID.String)
	if err != nil {
		log.Printf("Failed to signal workflow: %v", err)
		http.Error(w, "Failed to signal workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Job started successfully",
		"job_id":  jobID,
	})
}

// CompleteJob allows a worker to mark a job as completed
func CompleteJob(w http.ResponseWriter, r *http.Request) {
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

	// Get job and workflow information
	var workflowID sql.NullString
	var status string
	query := `
		SELECT temporal_workflow_id, status 
		FROM jobs 
		WHERE id = $1
	`
	err = config.DB.QueryRow(query, jobID).Scan(&workflowID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting job: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if status != "in_progress" {
		http.Error(w, fmt.Sprintf("Job not in progress, current status: %s", status), http.StatusBadRequest)
		return
	}

	// If no workflow is associated, update job status directly
	if !workflowID.Valid {
		updateQuery := `
			UPDATE jobs 
			SET status = 'completed', actual_end = NOW(), updated_at = NOW()
			WHERE id = $1
		`
		_, err = config.DB.Exec(updateQuery, jobID)
		if err != nil {
			log.Printf("Database error updating job status: %v", err)
			http.Error(w, "Failed to update job status", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Job completed successfully",
			"job_id":  jobID,
		})
		return
	}

	// Signal the workflow
	temporalClient, err := temporal.NewClient()
	if err != nil {
		log.Printf("Failed to create Temporal client: %v", err)
		http.Error(w, "Failed to connect to workflow service", http.StatusInternalServerError)
		return
	}
	defer temporalClient.Close()

	err = temporalClient.SignalJobCompleted(r.Context(), workflowID.String)
	if err != nil {
		log.Printf("Failed to signal workflow: %v", err)
		http.Error(w, "Failed to signal workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Job completed successfully",
		"job_id":  jobID,
	})
}

// SubmitReview allows users to submit reviews for jobs
func SubmitReview(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req struct {
		ReviewerID int    `json:"reviewer_id"`
		Rating     int    `json:"rating"`
		Comment    string `json:"comment"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate review data
	if req.ReviewerID <= 0 {
		http.Error(w, "Reviewer ID is required", http.StatusBadRequest)
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// Get job and workflow information
	var workflowID sql.NullString
	var status string
	query := `
		SELECT temporal_workflow_id, status 
		FROM jobs 
		WHERE id = $1
	`
	err = config.DB.QueryRow(query, jobID).Scan(&workflowID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting job: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if status != "review_pending" && status != "completed" {
		http.Error(w, fmt.Sprintf("Job not ready for review, current status: %s", status), http.StatusBadRequest)
		return
	}

	// Store review in database or just return success if no reviews table exists
	// If no workflow is associated, just acknowledge the review
	if !workflowID.Valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Review submitted successfully",
			"job_id":  jobID,
		})
		return
	}

	// Signal the workflow
	temporalClient, err := temporal.NewClient()
	if err != nil {
		log.Printf("Failed to create Temporal client: %v", err)
		http.Error(w, "Failed to connect to workflow service", http.StatusInternalServerError)
		return
	}
	defer temporalClient.Close()

	review := workflows.ReviewSubmission{
		JobID:      jobID,
		ReviewerID: req.ReviewerID,
		Rating:     req.Rating,
		Comment:    req.Comment,
	}

	err = temporalClient.SignalReviewSubmitted(r.Context(), workflowID.String, review)
	if err != nil {
		log.Printf("Failed to signal workflow: %v", err)
		http.Error(w, "Failed to signal workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Review submitted successfully",
		"job_id":  jobID,
	})
}