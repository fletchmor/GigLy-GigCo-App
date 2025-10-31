package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ==============================================
// PAYMENT ENUMS
// ==============================================

type TransactionType string

const (
	TransactionTypeAuthorization TransactionType = "authorization"
	TransactionTypeCapture       TransactionType = "capture"
	TransactionTypeCharge        TransactionType = "charge"
	TransactionTypeRefund        TransactionType = "refund"
	TransactionTypeVoid          TransactionType = "void"
	TransactionTypeAdjustment    TransactionType = "adjustment"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusRefunded  TransactionStatus = "refunded"
)

type PaymentSplitType string

const (
	PaymentSplitTypePlatformFee   PaymentSplitType = "platform_fee"
	PaymentSplitTypeWorkerPayment PaymentSplitType = "worker_payment"
	PaymentSplitTypeTax           PaymentSplitType = "tax"
	PaymentSplitTypeTip           PaymentSplitType = "tip"
	PaymentSplitTypeOther         PaymentSplitType = "other"
)

// ==============================================
// PAYMENT MODELS
// ==============================================

// EnhancedTransaction extends the basic Transaction with Clover fields
type EnhancedTransaction struct {
	ID                       int                `json:"id"`
	UUID                     string             `json:"uuid"`
	JobID                    int                `json:"job_id"`
	ConsumerID               int                `json:"consumer_id"`
	GigWorkerID              *int               `json:"gig_worker_id,omitempty"`
	Amount                   float64            `json:"amount"`
	Currency                 string             `json:"currency"`
	Status                   TransactionStatus  `json:"status"`
	TransactionType          TransactionType    `json:"transaction_type"`
	CloverChargeID           *string            `json:"clover_charge_id,omitempty"`
	CloverPaymentID          *string            `json:"clover_payment_id,omitempty"`
	CloverSourceToken        *string            `json:"clover_source_token,omitempty"`
	CloverRefundID           *string            `json:"clover_refund_id,omitempty"`
	CloverOrderID            *string            `json:"clover_order_id,omitempty"`
	AuthorizedAt             *time.Time         `json:"authorized_at,omitempty"`
	AuthorizationExpiresAt   *time.Time         `json:"authorization_expires_at,omitempty"`
	CapturedAt               *time.Time         `json:"captured_at,omitempty"`
	CaptureAmount            *float64           `json:"capture_amount,omitempty"`
	PaymentMethodID          *int               `json:"payment_method_id,omitempty"`
	PaymentMethod            *string            `json:"payment_method,omitempty"`
	LastFour                 *string            `json:"last_four,omitempty"`
	ProcessingFee            float64            `json:"processing_fee"`
	PlatformFee              float64            `json:"platform_fee"`
	NetAmount                *float64           `json:"net_amount,omitempty"`
	EscrowHeldAt             *time.Time         `json:"escrow_held_at,omitempty"`
	EscrowReleasedAt         *time.Time         `json:"escrow_released_at,omitempty"`
	RefundedAt               *time.Time         `json:"refunded_at,omitempty"`
	RefundAmount             *float64           `json:"refund_amount,omitempty"`
	RefundReason             *string            `json:"refund_reason,omitempty"`
	SettlementBatchID        *int               `json:"settlement_batch_id,omitempty"`
	ReconciledAt             *time.Time         `json:"reconciled_at,omitempty"`
	ParentTransactionID      *int               `json:"parent_transaction_id,omitempty"`
	Metadata                 *JSONB             `json:"metadata,omitempty"`
	Notes                    *string            `json:"notes,omitempty"`
	FailureReason            *string            `json:"failure_reason,omitempty"`
	CreatedAt                time.Time          `json:"created_at"`
	UpdatedAt                time.Time          `json:"updated_at"`
	Splits                   []PaymentSplit     `json:"splits,omitempty"`
}

type PaymentSplit struct {
	ID            int              `json:"id"`
	UUID          string           `json:"uuid"`
	TransactionID int              `json:"transaction_id"`
	SplitType     PaymentSplitType `json:"split_type"`
	Amount        float64          `json:"amount"`
	Percentage    *float64         `json:"percentage,omitempty"`
	RecipientID   *int             `json:"recipient_id,omitempty"`
	Description   *string          `json:"description,omitempty"`
	Metadata      *JSONB           `json:"metadata,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type PaymentEvent struct {
	ID              int       `json:"id"`
	UUID            string    `json:"uuid"`
	TransactionID   int       `json:"transaction_id"`
	EventType       string    `json:"event_type"`
	EventStatus     string    `json:"event_status"`
	CloverResponse  *JSONB    `json:"clover_response,omitempty"`
	ErrorMessage    *string   `json:"error_message,omitempty"`
	ErrorCode       *string   `json:"error_code,omitempty"`
	IdempotencyKey  *string   `json:"idempotency_key,omitempty"`
	UserID          *int      `json:"user_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type UserPaymentMethod struct {
	ID                int       `json:"id"`
	UUID              string    `json:"uuid"`
	UserID            int       `json:"user_id"`
	ProviderID        int       `json:"provider_id"`
	ExternalID        string    `json:"external_id"`
	Type              string    `json:"type"`
	LastFour          *string   `json:"last_four,omitempty"`
	Brand             *string   `json:"brand,omitempty"`
	IsDefault         bool      `json:"is_default"`
	IsActive          bool      `json:"is_active"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	CloverToken       *string   `json:"clover_token,omitempty"`
	CloverCustomerID  *string   `json:"clover_customer_id,omitempty"`
	Fingerprint       *string   `json:"fingerprint,omitempty"`
	Metadata          *JSONB    `json:"metadata,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ==============================================
// CLOVER API MODELS
// ==============================================

type CloverConfig struct {
	Environment          string `json:"environment"`           // sandbox or production
	MerchantID           string `json:"merchant_id"`
	AccessToken          string `json:"access_token"`
	APIAccessKey         string `json:"api_access_key"`         // PAKMS key for tokenization
	TokenizationEndpoint string `json:"tokenization_endpoint"`
	APIEndpoint          string `json:"api_endpoint"`
	WebhookSecret        string `json:"webhook_secret,omitempty"`
}

// Clover Tokenization Request
type CloverTokenizeRequest struct {
	Card CloverCard `json:"card"`
}

type CloverCard struct {
	Number      string `json:"number"`
	ExpMonth    string `json:"exp_month"`
	ExpYear     string `json:"exp_year"`
	CVV         string `json:"cvv"`
	Brand       string `json:"brand,omitempty"`       // visa, mastercard, etc
	Name        string `json:"name,omitempty"`        // cardholder name
	AddressLine1 string `json:"address_line1,omitempty"`
	AddressLine2 string `json:"address_line2,omitempty"`
	AddressCity  string `json:"address_city,omitempty"`
	AddressState string `json:"address_state,omitempty"`
	AddressZip   string `json:"address_zip,omitempty"`
	AddressCountry string `json:"address_country,omitempty"`
}

// Clover Tokenization Response
type CloverTokenizeResponse struct {
	ID       string `json:"id"`        // Token ID (clv_xxx)
	Object   string `json:"object"`
	Card     CloverTokenCard `json:"card"`
}

type CloverTokenCard struct {
	Brand       string `json:"brand"`
	ExpMonth    string `json:"exp_month"`
	ExpYear     string `json:"exp_year"`
	First6      string `json:"first6"`
	Last4       string `json:"last4"`
}

// Clover Charge Request (Authorization or Direct Charge)
type CloverChargeRequest struct {
	Amount          int64   `json:"amount"`            // Amount in cents
	Currency        string  `json:"currency"`          // USD
	Source          string  `json:"source"`            // Token ID
	Capture         bool    `json:"capture"`           // false for pre-auth
	Description     string  `json:"description,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ExternalPaymentID string `json:"ecomind,omitempty"` // External reference ID
}

// Clover Charge Response
type CloverChargeResponse struct {
	ID              string                 `json:"id"`
	Amount          int64                  `json:"amount"`
	Currency        string                 `json:"currency"`
	Created         int64                  `json:"created"`
	Captured        bool                   `json:"captured"`
	Status          string                 `json:"status"`          // succeeded, failed, pending
	Source          CloverSourceResponse   `json:"source"`
	Outcome         *CloverOutcome         `json:"outcome,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	RefundedAmount  int64                  `json:"amount_refunded,omitempty"`
	FailureCode     string                 `json:"failure_code,omitempty"`
	FailureMessage  string                 `json:"failure_message,omitempty"`
}

type CloverSourceResponse struct {
	ID            string `json:"id"`
	Brand         string `json:"brand"`
	Last4         string `json:"last4"`
	ExpMonth      string `json:"exp_month"`
	ExpYear       string `json:"exp_year"`
	Fingerprint   string `json:"fingerprint,omitempty"`
}

type CloverOutcome struct {
	NetworkStatus string `json:"network_status"`
	Reason        string `json:"reason,omitempty"`
	RiskLevel     string `json:"risk_level,omitempty"`
	SellerMessage string `json:"seller_message,omitempty"`
	Type          string `json:"type"`
}

// Clover Capture Request
type CloverCaptureRequest struct {
	Amount int64 `json:"amount,omitempty"` // Amount in cents, omit to capture full auth
}

// Clover Capture Response
type CloverCaptureResponse struct {
	ID             string `json:"id"`
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	Created        int64  `json:"created"`
	Status         string `json:"status"`
	PaymentID      string `json:"payment_id"`
}

// Clover Refund Request
type CloverRefundRequest struct {
	ChargeID string `json:"charge,omitempty"` // Charge ID to refund
	Amount   int64  `json:"amount,omitempty"` // Amount in cents, omit for full refund
	Reason   string `json:"reason,omitempty"`
}

// Clover Refund Response
type CloverRefundResponse struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Created  int64  `json:"created"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	ChargeID string `json:"charge"`
	Reason   string `json:"reason,omitempty"`
}

// ==============================================
// API REQUEST/RESPONSE MODELS
// ==============================================

// Payment authorization request
type PaymentAuthorizeRequest struct {
	JobID             int                 `json:"job_id" binding:"required"`
	PaymentMethodID   *int                `json:"payment_method_id,omitempty"`
	CardToken         *string             `json:"card_token,omitempty"`
	CardDetails       *CardDetails        `json:"card_details,omitempty"`
	Amount            float64             `json:"amount" binding:"required,gt=0"`
	SaveCard          bool                `json:"save_card"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type CardDetails struct {
	Number      string `json:"number" binding:"required"`
	ExpMonth    string `json:"exp_month" binding:"required"`
	ExpYear     string `json:"exp_year" binding:"required"`
	CVV         string `json:"cvv" binding:"required"`
	Name        string `json:"name,omitempty"`
	AddressLine1 string `json:"address_line1,omitempty"`
	AddressCity  string `json:"address_city,omitempty"`
	AddressState string `json:"address_state,omitempty"`
	AddressZip   string `json:"address_zip,omitempty"`
}

type PaymentAuthorizeResponse struct {
	Success       bool                `json:"success"`
	TransactionID int                 `json:"transaction_id"`
	Transaction   *EnhancedTransaction `json:"transaction,omitempty"`
	Message       string              `json:"message,omitempty"`
}

// Payment capture request
type PaymentCaptureRequest struct {
	TransactionID int      `json:"transaction_id" binding:"required"`
	Amount        *float64 `json:"amount,omitempty"` // Omit for full capture
}

type PaymentCaptureResponse struct {
	Success       bool                `json:"success"`
	TransactionID int                 `json:"transaction_id"`
	Transaction   *EnhancedTransaction `json:"transaction,omitempty"`
	Message       string              `json:"message,omitempty"`
}

// Payment refund request
type PaymentRefundRequest struct {
	TransactionID int      `json:"transaction_id" binding:"required"`
	Amount        *float64 `json:"amount,omitempty"` // Omit for full refund
	Reason        string   `json:"reason,omitempty"`
}

type PaymentRefundResponse struct {
	Success       bool                `json:"success"`
	RefundID      int                 `json:"refund_id"`
	Transaction   *EnhancedTransaction `json:"transaction,omitempty"`
	Message       string              `json:"message,omitempty"`
}

// Payment method save request
type SavePaymentMethodRequest struct {
	CardDetails   CardDetails `json:"card_details" binding:"required"`
	IsDefault     bool        `json:"is_default"`
}

type SavePaymentMethodResponse struct {
	Success        bool              `json:"success"`
	PaymentMethod  *UserPaymentMethod `json:"payment_method,omitempty"`
	Message        string            `json:"message,omitempty"`
}

// Job payment summary
type JobPaymentSummary struct {
	JobID            int     `json:"job_id"`
	TotalAuthorized  float64 `json:"total_authorized"`
	TotalCaptured    float64 `json:"total_captured"`
	TotalRefunded    float64 `json:"total_refunded"`
	PlatformFees     float64 `json:"platform_fees"`
	WorkerPayment    float64 `json:"worker_payment"`
	EscrowStatus     string  `json:"escrow_status"` // held, released, none
}

// ==============================================
// JSONB TYPE FOR POSTGRES
// ==============================================

type JSONB map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	result := make(map[string]interface{})
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
