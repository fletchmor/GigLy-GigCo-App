package model

import (
	"time"
)

// Review represents a job review/rating
type Review struct {
	ID         int       `json:"id" db:"id"`
	UUID       string    `json:"uuid" db:"uuid"`
	JobID      int       `json:"job_id" db:"job_id"`
	ReviewerID int       `json:"reviewer_id" db:"reviewer_id"`
	RevieweeID int       `json:"reviewee_id" db:"reviewee_id"`
	Rating     int       `json:"rating" db:"rating"`
	ReviewText *string   `json:"review_text" db:"review_text"`
	IsPublic   bool      `json:"is_public" db:"is_public"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// ReviewWithDetails includes reviewer and reviewee information
type ReviewWithDetails struct {
	Review
	ReviewerName string  `json:"reviewer_name" db:"reviewer_name"`
	RevieweeName string  `json:"reviewee_name" db:"reviewee_name"`
	JobTitle     string  `json:"job_title" db:"job_title"`
	JobCategory  *string `json:"job_category" db:"job_category"`
}

// ReviewRequest represents the request payload for creating a review
type ReviewRequest struct {
	JobID      int     `json:"job_id" validate:"required,min=1"`
	ReviewerID int     `json:"reviewer_id" validate:"required,min=1"`
	RevieweeID int     `json:"reviewee_id" validate:"required,min=1"`
	Rating     int     `json:"rating" validate:"required,min=1,max=5"`
	ReviewText *string `json:"review_text" validate:"omitempty,max=1000"`
	IsPublic   *bool   `json:"is_public"`
}

// ReviewUpdateRequest represents the request payload for updating a review
type ReviewUpdateRequest struct {
	Rating     *int    `json:"rating" validate:"omitempty,min=1,max=5"`
	ReviewText *string `json:"review_text" validate:"omitempty,max=1000"`
	IsPublic   *bool   `json:"is_public"`
}

// ReviewStats represents aggregated review statistics
type ReviewStats struct {
	UserID         int     `json:"user_id" db:"user_id"`
	UserName       string  `json:"user_name" db:"user_name"`
	UserRole       string  `json:"user_role" db:"user_role"`
	TotalReviews   int     `json:"total_reviews" db:"total_reviews"`
	AverageRating  float64 `json:"average_rating" db:"average_rating"`
	Rating5Count   int     `json:"rating_5_count" db:"rating_5_count"`
	Rating4Count   int     `json:"rating_4_count" db:"rating_4_count"`
	Rating3Count   int     `json:"rating_3_count" db:"rating_3_count"`
	Rating2Count   int     `json:"rating_2_count" db:"rating_2_count"`
	Rating1Count   int     `json:"rating_1_count" db:"rating_1_count"`
	LastReviewDate *time.Time `json:"last_review_date" db:"last_review_date"`
}

// ReviewFilters represents filtering options for review queries
type ReviewFilters struct {
	UserID        *int    `json:"user_id"`
	JobID         *int    `json:"job_id"`
	ReviewerID    *int    `json:"reviewer_id"`
	RevieweeID    *int    `json:"reviewee_id"`
	MinRating     *int    `json:"min_rating"`
	MaxRating     *int    `json:"max_rating"`
	IsPublic      *bool   `json:"is_public"`
	Category      *string `json:"category"`
	DateFrom      *string `json:"date_from"`
	DateTo        *string `json:"date_to"`
	Page          int     `json:"page"`
	Limit         int     `json:"limit"`
	SortBy        string  `json:"sort_by"` // "created_at", "rating", "job_title"
	SortOrder     string  `json:"sort_order"` // "asc", "desc"
}

// PaginatedReviews represents paginated review results
type PaginatedReviews struct {
	Reviews     []ReviewWithDetails `json:"reviews"`
	Pagination  Pagination          `json:"pagination"`
}

// GetDefaultFilters returns default filter values
func (f *ReviewFilters) GetDefaultFilters() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 || f.Limit > 50 {
		f.Limit = 20
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}
}

// ValidateRating ensures the rating is within valid range
func ValidateRating(rating int) bool {
	return rating >= 1 && rating <= 5
}

// CalculateAverageRating calculates average rating from a distribution
func CalculateAverageRating(rating1, rating2, rating3, rating4, rating5 int) float64 {
	total := rating1 + rating2 + rating3 + rating4 + rating5
	if total == 0 {
		return 0.0
	}
	
	weightedSum := (rating1 * 1) + (rating2 * 2) + (rating3 * 3) + (rating4 * 4) + (rating5 * 5)
	return float64(weightedSum) / float64(total)
}