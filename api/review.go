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

// CreateReview allows users to submit a review for a completed job
func CreateReview(w http.ResponseWriter, r *http.Request) {
	var req model.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.JobID <= 0 {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}
	if req.ReviewerID <= 0 {
		http.Error(w, "Reviewer ID is required", http.StatusBadRequest)
		return
	}
	if req.RevieweeID <= 0 {
		http.Error(w, "Reviewee ID is required", http.StatusBadRequest)
		return
	}
	if !model.ValidateRating(req.Rating) {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}
	if req.ReviewerID == req.RevieweeID {
		http.Error(w, "Cannot review yourself", http.StatusBadRequest)
		return
	}

	// Set default visibility
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	// Check if job exists and is completed
	var jobStatus string
	var consumerID, gigWorkerID sql.NullInt32
	jobQuery := `
		SELECT status, consumer_id, gig_worker_id
		FROM jobs 
		WHERE id = $1
	`
	err := config.DB.QueryRow(jobQuery, req.JobID).Scan(&jobStatus, &consumerID, &gigWorkerID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting job: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Ensure job is completed
	if jobStatus != "completed" {
		http.Error(w, "Job must be completed before submitting a review", http.StatusBadRequest)
		return
	}

	// Validate that reviewer and reviewee are part of this job
	validReviewer := (consumerID.Valid && int(consumerID.Int32) == req.ReviewerID) || 
					 (gigWorkerID.Valid && int(gigWorkerID.Int32) == req.ReviewerID)
	validReviewee := (consumerID.Valid && int(consumerID.Int32) == req.RevieweeID) || 
					 (gigWorkerID.Valid && int(gigWorkerID.Int32) == req.RevieweeID)

	if !validReviewer || !validReviewee {
		http.Error(w, "Reviewer and reviewee must be participants in this job", http.StatusBadRequest)
		return
	}

	// Check if review already exists
	var existingID int
	checkQuery := `SELECT id FROM job_reviews WHERE job_id = $1 AND reviewer_id = $2`
	err = config.DB.QueryRow(checkQuery, req.JobID, req.ReviewerID).Scan(&existingID)
	if err == nil {
		http.Error(w, "Review already exists for this job", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		log.Printf("Database error checking existing review: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert new review
	insertQuery := `
		INSERT INTO job_reviews (job_id, reviewer_id, reviewee_id, rating, review_text, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, uuid, created_at, updated_at
	`
	
	var review model.Review
	review.JobID = req.JobID
	review.ReviewerID = req.ReviewerID
	review.RevieweeID = req.RevieweeID
	review.Rating = req.Rating
	review.ReviewText = req.ReviewText
	review.IsPublic = isPublic

	err = config.DB.QueryRow(insertQuery, req.JobID, req.ReviewerID, req.RevieweeID, req.Rating, req.ReviewText, isPublic).
		Scan(&review.ID, &review.UUID, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		log.Printf("Database error creating review: %v", err)
		http.Error(w, "Failed to create review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Review created successfully",
		"review":  review,
	})
}

// GetReviews retrieves reviews with filtering and pagination
func GetReviews(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters := model.ReviewFilters{
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters.Page = p
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 50 {
			filters.Limit = l
		}
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		if uid, err := strconv.Atoi(userID); err == nil {
			filters.UserID = &uid
		}
	}
	if jobID := r.URL.Query().Get("job_id"); jobID != "" {
		if jid, err := strconv.Atoi(jobID); err == nil {
			filters.JobID = &jid
		}
	}
	if reviewerID := r.URL.Query().Get("reviewer_id"); reviewerID != "" {
		if rid, err := strconv.Atoi(reviewerID); err == nil {
			filters.ReviewerID = &rid
		}
	}
	if revieweeID := r.URL.Query().Get("reviewee_id"); revieweeID != "" {
		if rid, err := strconv.Atoi(revieweeID); err == nil {
			filters.RevieweeID = &rid
		}
	}
	if minRating := r.URL.Query().Get("min_rating"); minRating != "" {
		if mr, err := strconv.Atoi(minRating); err == nil && mr >= 1 && mr <= 5 {
			filters.MinRating = &mr
		}
	}
	if maxRating := r.URL.Query().Get("max_rating"); maxRating != "" {
		if mr, err := strconv.Atoi(maxRating); err == nil && mr >= 1 && mr <= 5 {
			filters.MaxRating = &mr
		}
	}
	if isPublic := r.URL.Query().Get("is_public"); isPublic != "" {
		if ip, err := strconv.ParseBool(isPublic); err == nil {
			filters.IsPublic = &ip
		}
	}
	if category := r.URL.Query().Get("category"); category != "" {
		filters.Category = &category
	}
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		filters.DateFrom = &dateFrom
	}
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		filters.DateTo = &dateTo
	}
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		validSortBy := []string{"created_at", "rating", "job_title"}
		for _, valid := range validSortBy {
			if sortBy == valid {
				filters.SortBy = sortBy
				break
			}
		}
	}
	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder == "asc" || sortOrder == "desc" {
		filters.SortOrder = sortOrder
	}

	// Build the query
	baseQuery := `
		SELECT 
			r.id, r.uuid, r.job_id, r.reviewer_id, r.reviewee_id, 
			r.rating, r.review_text, r.is_public, r.created_at, r.updated_at,
			reviewer.name as reviewer_name,
			reviewee.name as reviewee_name,
			j.title as job_title,
			j.category as job_category
		FROM job_reviews r
		JOIN people reviewer ON reviewer.id = r.reviewer_id
		JOIN people reviewee ON reviewee.id = r.reviewee_id
		JOIN jobs j ON j.id = r.job_id
		WHERE 1=1
	`
	
	countQuery := `
		SELECT COUNT(*)
		FROM job_reviews r
		JOIN people reviewer ON reviewer.id = r.reviewer_id
		JOIN people reviewee ON reviewee.id = r.reviewee_id
		JOIN jobs j ON j.id = r.job_id
		WHERE 1=1
	`

	var args []interface{}
	var whereConditions []string
	argIndex := 1

	// Apply filters
	if filters.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("(r.reviewer_id = $%d OR r.reviewee_id = $%d)", argIndex, argIndex+1))
		args = append(args, *filters.UserID, *filters.UserID)
		argIndex += 2
	}
	if filters.JobID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.job_id = $%d", argIndex))
		args = append(args, *filters.JobID)
		argIndex++
	}
	if filters.ReviewerID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.reviewer_id = $%d", argIndex))
		args = append(args, *filters.ReviewerID)
		argIndex++
	}
	if filters.RevieweeID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.reviewee_id = $%d", argIndex))
		args = append(args, *filters.RevieweeID)
		argIndex++
	}
	if filters.MinRating != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.rating >= $%d", argIndex))
		args = append(args, *filters.MinRating)
		argIndex++
	}
	if filters.MaxRating != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.rating <= $%d", argIndex))
		args = append(args, *filters.MaxRating)
		argIndex++
	}
	if filters.IsPublic != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.is_public = $%d", argIndex))
		args = append(args, *filters.IsPublic)
		argIndex++
	}
	if filters.Category != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("j.category = $%d", argIndex))
		args = append(args, *filters.Category)
		argIndex++
	}
	if filters.DateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.created_at >= $%d", argIndex))
		args = append(args, *filters.DateFrom)
		argIndex++
	}
	if filters.DateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.created_at <= $%d", argIndex))
		args = append(args, *filters.DateTo)
		argIndex++
	}

	// Add WHERE conditions
	if len(whereConditions) > 0 {
		whereClause := " AND " + strings.Join(whereConditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Only show public reviews unless specifically filtered
	if filters.IsPublic == nil {
		baseQuery += " AND r.is_public = true"
		countQuery += " AND r.is_public = true"
	}

	// Get total count
	var totalCount int
	err := config.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		log.Printf("Database error getting review count: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add sorting and pagination
	baseQuery += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", 
		filters.SortBy, strings.ToUpper(filters.SortOrder), argIndex, argIndex+1)
	args = append(args, filters.Limit, (filters.Page-1)*filters.Limit)

	// Execute query
	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		log.Printf("Database error getting reviews: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviews []model.ReviewWithDetails
	for rows.Next() {
		var review model.ReviewWithDetails
		err := rows.Scan(
			&review.ID, &review.UUID, &review.JobID, &review.ReviewerID, &review.RevieweeID,
			&review.Rating, &review.ReviewText, &review.IsPublic, &review.CreatedAt, &review.UpdatedAt,
			&review.ReviewerName, &review.RevieweeName, &review.JobTitle, &review.JobCategory,
		)
		if err != nil {
			log.Printf("Error scanning review row: %v", err)
			continue
		}
		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create pagination info
	totalPages := (totalCount + filters.Limit - 1) / filters.Limit
	pagination := model.Pagination{
		Page:    filters.Page,
		Limit:   filters.Limit,
		Total:   totalCount,
		Pages:   totalPages,
		HasNext: filters.Page < totalPages,
		HasPrev: filters.Page > 1,
	}

	response := model.PaginatedReviews{
		Reviews:    reviews,
		Pagination: pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetReviewByID retrieves a specific review by ID
func GetReviewByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	reviewID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid review ID format", http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
			r.id, r.uuid, r.job_id, r.reviewer_id, r.reviewee_id, 
			r.rating, r.review_text, r.is_public, r.created_at, r.updated_at,
			reviewer.name as reviewer_name,
			reviewee.name as reviewee_name,
			j.title as job_title,
			j.category as job_category
		FROM job_reviews r
		JOIN people reviewer ON reviewer.id = r.reviewer_id
		JOIN people reviewee ON reviewee.id = r.reviewee_id
		JOIN jobs j ON j.id = r.job_id
		WHERE r.id = $1
	`

	var review model.ReviewWithDetails
	err = config.DB.QueryRow(query, reviewID).Scan(
		&review.ID, &review.UUID, &review.JobID, &review.ReviewerID, &review.RevieweeID,
		&review.Rating, &review.ReviewText, &review.IsPublic, &review.CreatedAt, &review.UpdatedAt,
		&review.ReviewerName, &review.RevieweeName, &review.JobTitle, &review.JobCategory,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting review: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Only return public reviews unless specifically authorized
	if !review.IsPublic {
		// Here you could add authorization logic to check if the current user
		// is the reviewer, reviewee, or an admin
		http.Error(w, "Review is private", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(review)
}

// UpdateReview allows updating an existing review
func UpdateReview(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	reviewID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid review ID format", http.StatusBadRequest)
		return
	}

	var req model.ReviewUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Check if review exists
	var existingReview model.Review
	checkQuery := `SELECT id, reviewer_id, reviewee_id FROM job_reviews WHERE id = $1`
	err = config.DB.QueryRow(checkQuery, reviewID).Scan(&existingReview.ID, &existingReview.ReviewerID, &existingReview.RevieweeID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting review: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Here you could add authorization logic to ensure the current user
	// is the original reviewer

	// Build update query dynamically
	var updateParts []string
	var args []interface{}
	argIndex := 1

	if req.Rating != nil {
		if !model.ValidateRating(*req.Rating) {
			http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
			return
		}
		updateParts = append(updateParts, fmt.Sprintf("rating = $%d", argIndex))
		args = append(args, *req.Rating)
		argIndex++
	}
	if req.ReviewText != nil {
		updateParts = append(updateParts, fmt.Sprintf("review_text = $%d", argIndex))
		args = append(args, *req.ReviewText)
		argIndex++
	}
	if req.IsPublic != nil {
		updateParts = append(updateParts, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *req.IsPublic)
		argIndex++
	}

	if len(updateParts) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Add updated_at
	updateParts = append(updateParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, reviewID)
	updateQuery := fmt.Sprintf("UPDATE job_reviews SET %s WHERE id = $%d", strings.Join(updateParts, ", "), argIndex)

	_, err = config.DB.Exec(updateQuery, args...)
	if err != nil {
		log.Printf("Database error updating review: %v", err)
		http.Error(w, "Failed to update review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Review updated successfully",
	})
}

// DeleteReview allows deleting a review
func DeleteReview(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	reviewID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid review ID format", http.StatusBadRequest)
		return
	}

	// Check if review exists
	var reviewerID int
	checkQuery := `SELECT reviewer_id FROM job_reviews WHERE id = $1`
	err = config.DB.QueryRow(checkQuery, reviewID).Scan(&reviewerID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting review: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Here you could add authorization logic to ensure the current user
	// is the original reviewer or an admin

	// Delete review
	deleteQuery := `DELETE FROM job_reviews WHERE id = $1`
	_, err = config.DB.Exec(deleteQuery, reviewID)
	if err != nil {
		log.Printf("Database error deleting review: %v", err)
		http.Error(w, "Failed to delete review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Review deleted successfully",
	})
}

// GetUserReviewStats retrieves aggregated review statistics for a user
func GetUserReviewStats(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
			p.id as user_id,
			p.name as user_name,
			p.role as user_role,
			COUNT(r.id) as total_reviews,
			COALESCE(AVG(r.rating::numeric), 0) as average_rating,
			COUNT(CASE WHEN r.rating = 5 THEN 1 END) as rating_5_count,
			COUNT(CASE WHEN r.rating = 4 THEN 1 END) as rating_4_count,
			COUNT(CASE WHEN r.rating = 3 THEN 1 END) as rating_3_count,
			COUNT(CASE WHEN r.rating = 2 THEN 1 END) as rating_2_count,
			COUNT(CASE WHEN r.rating = 1 THEN 1 END) as rating_1_count,
			MAX(r.created_at) as last_review_date
		FROM people p
		LEFT JOIN job_reviews r ON r.reviewee_id = p.id AND r.is_public = true
		WHERE p.id = $1 AND p.is_active = true
		GROUP BY p.id, p.name, p.role
	`

	var stats model.ReviewStats
	err = config.DB.QueryRow(query, userID).Scan(
		&stats.UserID, &stats.UserName, &stats.UserRole, &stats.TotalReviews,
		&stats.AverageRating, &stats.Rating5Count, &stats.Rating4Count,
		&stats.Rating3Count, &stats.Rating2Count, &stats.Rating1Count,
		&stats.LastReviewDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting user review stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Round average rating to 2 decimal places
	stats.AverageRating = float64(int(stats.AverageRating*100)) / 100

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetJobReviews retrieves all reviews for a specific job
func GetJobReviews(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	jobID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid job ID format", http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
			r.id, r.uuid, r.job_id, r.reviewer_id, r.reviewee_id, 
			r.rating, r.review_text, r.is_public, r.created_at, r.updated_at,
			reviewer.name as reviewer_name,
			reviewee.name as reviewee_name,
			j.title as job_title,
			j.category as job_category
		FROM job_reviews r
		JOIN people reviewer ON reviewer.id = r.reviewer_id
		JOIN people reviewee ON reviewee.id = r.reviewee_id
		JOIN jobs j ON j.id = r.job_id
		WHERE r.job_id = $1 AND r.is_public = true
		ORDER BY r.created_at DESC
	`

	rows, err := config.DB.Query(query, jobID)
	if err != nil {
		log.Printf("Database error getting job reviews: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviews []model.ReviewWithDetails
	for rows.Next() {
		var review model.ReviewWithDetails
		err := rows.Scan(
			&review.ID, &review.UUID, &review.JobID, &review.ReviewerID, &review.RevieweeID,
			&review.Rating, &review.ReviewText, &review.IsPublic, &review.CreatedAt, &review.UpdatedAt,
			&review.ReviewerName, &review.RevieweeName, &review.JobTitle, &review.JobCategory,
		)
		if err != nil {
			log.Printf("Error scanning review row: %v", err)
			continue
		}
		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":  jobID,
		"reviews": reviews,
	})
}

// GetPlatformReviewStats retrieves platform-wide review statistics
func GetPlatformReviewStats(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT 
			COUNT(*) as total_reviews,
			COALESCE(AVG(rating::numeric), 0) as average_rating,
			COUNT(CASE WHEN rating = 5 THEN 1 END) as rating_5_count,
			COUNT(CASE WHEN rating = 4 THEN 1 END) as rating_4_count,
			COUNT(CASE WHEN rating = 3 THEN 1 END) as rating_3_count,
			COUNT(CASE WHEN rating = 2 THEN 1 END) as rating_2_count,
			COUNT(CASE WHEN rating = 1 THEN 1 END) as rating_1_count,
			COUNT(DISTINCT reviewee_id) as reviewed_users,
			MAX(created_at) as latest_review_date,
			MIN(created_at) as first_review_date
		FROM job_reviews 
		WHERE is_public = true
	`

	var stats struct {
		TotalReviews      int        `json:"total_reviews"`
		AverageRating     float64    `json:"average_rating"`
		Rating5Count      int        `json:"rating_5_count"`
		Rating4Count      int        `json:"rating_4_count"`
		Rating3Count      int        `json:"rating_3_count"`
		Rating2Count      int        `json:"rating_2_count"`
		Rating1Count      int        `json:"rating_1_count"`
		ReviewedUsers     int        `json:"reviewed_users"`
		LatestReviewDate  *time.Time `json:"latest_review_date"`
		FirstReviewDate   *time.Time `json:"first_review_date"`
	}

	err := config.DB.QueryRow(query).Scan(
		&stats.TotalReviews, &stats.AverageRating,
		&stats.Rating5Count, &stats.Rating4Count, &stats.Rating3Count,
		&stats.Rating2Count, &stats.Rating1Count,
		&stats.ReviewedUsers, &stats.LatestReviewDate, &stats.FirstReviewDate,
	)
	if err != nil {
		log.Printf("Database error getting platform review stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Round average rating to 2 decimal places
	stats.AverageRating = float64(int(stats.AverageRating*100)) / 100

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetTopRatedUsers retrieves the highest rated users on the platform
func GetTopRatedUsers(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limit := 10
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Parse role filter
	roleFilter := r.URL.Query().Get("role")

	baseQuery := `
		SELECT 
			p.id as user_id,
			p.name as user_name,
			p.role as user_role,
			COUNT(r.id) as total_reviews,
			COALESCE(AVG(r.rating::numeric), 0) as average_rating,
			COUNT(CASE WHEN r.rating = 5 THEN 1 END) as rating_5_count,
			COUNT(CASE WHEN r.rating = 4 THEN 1 END) as rating_4_count,
			COUNT(CASE WHEN r.rating = 3 THEN 1 END) as rating_3_count,
			COUNT(CASE WHEN r.rating = 2 THEN 1 END) as rating_2_count,
			COUNT(CASE WHEN r.rating = 1 THEN 1 END) as rating_1_count,
			MAX(r.created_at) as last_review_date
		FROM people p
		LEFT JOIN job_reviews r ON r.reviewee_id = p.id AND r.is_public = true
		WHERE p.is_active = true
	`

	var args []interface{}
	argIndex := 1

	if roleFilter != "" {
		baseQuery += fmt.Sprintf(" AND p.role = $%d", argIndex)
		args = append(args, roleFilter)
		argIndex++
	}

	baseQuery += fmt.Sprintf(` 
		GROUP BY p.id, p.name, p.role
		HAVING COUNT(r.id) > 0
		ORDER BY average_rating DESC, total_reviews DESC
		LIMIT $%d
	`, argIndex)
	args = append(args, limit)

	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		log.Printf("Database error getting top rated users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var topUsers []model.ReviewStats
	for rows.Next() {
		var user model.ReviewStats
		err := rows.Scan(
			&user.UserID, &user.UserName, &user.UserRole, &user.TotalReviews,
			&user.AverageRating, &user.Rating5Count, &user.Rating4Count,
			&user.Rating3Count, &user.Rating2Count, &user.Rating1Count,
			&user.LastReviewDate,
		)
		if err != nil {
			log.Printf("Error scanning top rated user row: %v", err)
			continue
		}

		// Round average rating to 2 decimal places
		user.AverageRating = float64(int(user.AverageRating*100)) / 100
		topUsers = append(topUsers, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"top_rated_users": topUsers,
		"limit":          limit,
		"role_filter":    roleFilter,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}