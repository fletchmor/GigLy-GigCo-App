package model

import "time"

type User struct {
	ID            int       `json:"id"`
	Uuid          string    `json:"uuid"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	PlaceID       string    `json:"place_id"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Schedule struct {
	ID               int        `json:"id"`
	Uuid             string     `json:"uuid"`
	GigWorkerID      int        `json:"gig_worker_id"`
	Title            string     `json:"title"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          time.Time  `json:"end_time"`
	IsAvailable      bool       `json:"is_available"`
	JobID            *int       `json:"job_id"`
	RecurringPattern string     `json:"recurring_pattern"`
	RecurringUntil   *time.Time `json:"recurring_until"`
	Notes            string     `json:"notes"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Transaction struct {
	ID                int        `json:"id"`
	Uuid              string     `json:"uuid"`
	JobID             int        `json:"job_id"`
	ConsumerID        int        `json:"consumer_id"`
	GigWorkerID       int        `json:"gig_worker_id"`
	Amount            float64    `json:"amount"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	PaymentIntentID   string     `json:"payment_intent_id"`
	PaymentMethod     string     `json:"payment_method"`
	EscrowReleasedAt  *time.Time `json:"escrow_released_at"`
	ProcessingFee     float64    `json:"processing_fee"`
	NetAmount         float64    `json:"net_amount"`
	SettlementBatchID *int       `json:"settlement_batch_id"`
	Notes             string     `json:"notes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
