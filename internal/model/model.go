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

type GigWorker struct {
	ID                           int        `json:"id"`
	Uuid                         string     `json:"uuid"`
	Name                         string     `json:"name"`
	Email                        string     `json:"email"`
	Phone                        string     `json:"phone"`
	Address                      string     `json:"address"`
	Latitude                     float64    `json:"latitude"`
	Longitude                    float64    `json:"longitude"`
	PlaceID                      string     `json:"place_id"`
	Role                         string     `json:"role"`
	IsActive                     bool       `json:"is_active"`
	EmailVerified                bool       `json:"email_verified"`
	PhoneVerified                bool       `json:"phone_verified"`
	Bio                          string     `json:"bio,omitempty"`
	HourlyRate                   *float64   `json:"hourly_rate,omitempty"`
	ExperienceYears              *int       `json:"experience_years,omitempty"`
	VerificationStatus           string     `json:"verification_status,omitempty"`
	BackgroundCheckDate          *time.Time `json:"background_check_date,omitempty"`
	ServiceRadiusMiles           *float64   `json:"service_radius_miles,omitempty"`
	AvailabilityNotes            string     `json:"availability_notes,omitempty"`
	EmergencyContactName         string     `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone        string     `json:"emergency_contact_phone,omitempty"`
	EmergencyContactRelationship string     `json:"emergency_contact_relationship,omitempty"`
	CreatedAt                    time.Time  `json:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at"`
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

type Job struct {
	ID                     int        `json:"id"`
	UUID                   string     `json:"uuid"`
	ConsumerID             int        `json:"consumer_id"`
	GigWorkerID            *int       `json:"gig_worker_id,omitempty"`
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	Category               string     `json:"category,omitempty"`
	LocationAddress        string     `json:"location_address,omitempty"`
	LocationLatitude       *float64   `json:"location_latitude,omitempty"`
	LocationLongitude      *float64   `json:"location_longitude,omitempty"`
	EstimatedDurationHours *float64   `json:"estimated_duration_hours,omitempty"`
	PayRatePerHour         *float64   `json:"pay_rate_per_hour,omitempty"`
	TotalPay               *float64   `json:"total_pay,omitempty"`
	Status                 string     `json:"status"`
	ScheduledStart         *time.Time `json:"scheduled_start,omitempty"`
	ScheduledEnd           *time.Time `json:"scheduled_end,omitempty"`
	ActualStart            *time.Time `json:"actual_start,omitempty"`
	ActualEnd              *time.Time `json:"actual_end,omitempty"`
	Notes                  string     `json:"notes,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type JobCreateRequest struct {
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	Category               string     `json:"category,omitempty"`
	LocationAddress        string     `json:"location_address,omitempty"`
	Location               string     `json:"location,omitempty"` // Alternative for tests
	LocationLatitude       *float64   `json:"location_latitude,omitempty"`
	LocationLongitude      *float64   `json:"location_longitude,omitempty"`
	EstimatedDurationHours *float64   `json:"estimated_duration_hours,omitempty"`
	EstimatedHours         *float64   `json:"estimated_hours,omitempty"` // Alternative for tests
	PayRatePerHour         *float64   `json:"pay_rate_per_hour,omitempty"`
	PayRate                *float64   `json:"pay_rate,omitempty"` // Alternative for tests
	TotalPay               *float64   `json:"total_pay,omitempty"`
	ScheduledStart         *time.Time `json:"scheduled_start,omitempty"`
	ScheduledEnd           *time.Time `json:"scheduled_end,omitempty"`
	Notes                  string     `json:"notes,omitempty"`
	ConsumerID             int        `json:"consumer_id,omitempty"` // For tests
}

type JobUpdateRequest struct {
	Title                  *string    `json:"title,omitempty"`
	Description            *string    `json:"description,omitempty"`
	Category               *string    `json:"category,omitempty"`
	LocationAddress        *string    `json:"location_address,omitempty"`
	LocationLatitude       *float64   `json:"location_latitude,omitempty"`
	LocationLongitude      *float64   `json:"location_longitude,omitempty"`
	EstimatedDurationHours *float64   `json:"estimated_duration_hours,omitempty"`
	PayRatePerHour         *float64   `json:"pay_rate_per_hour,omitempty"`
	TotalPay               *float64   `json:"total_pay,omitempty"`
	ScheduledStart         *time.Time `json:"scheduled_start,omitempty"`
	ScheduledEnd           *time.Time `json:"scheduled_end,omitempty"`
	Notes                  *string    `json:"notes,omitempty"`
}

type JobResponse struct {
	Job
	Consumer  *UserSummary `json:"consumer,omitempty"`
	GigWorker *UserSummary `json:"gig_worker,omitempty"`
	Distance  *float64     `json:"distance_km,omitempty"`
}

type UserSummary struct {
	ID            int      `json:"id"`
	UUID          string   `json:"uuid"`
	Name          string   `json:"name"`
	AverageRating *float64 `json:"average_rating,omitempty"`
	TotalJobs     int      `json:"total_jobs,omitempty"`
}

type JobsListResponse struct {
	Jobs       []JobResponse `json:"jobs"`
	Pagination Pagination    `json:"pagination"`
}

type Pagination struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	Pages   int  `json:"pages"`
	HasNext bool `json:"has_next"`
	HasPrev bool `json:"has_prev"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
