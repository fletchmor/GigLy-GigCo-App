package api

import (
	"app/config"
	"app/internal/model"
	"app/internal/payment"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var paymentService *payment.PaymentService

// InitPaymentService initializes the payment service
func InitPaymentService() {
	if config.Payment == nil {
		config.InitPaymentConfig()
	}
	paymentService = payment.NewPaymentService(config.DB, &config.Payment.Clover)
	log.Println("Payment service initialized")
}

// ==============================================
// PAYMENT AUTHORIZATION (ESCROW)
// ==============================================

// AuthorizeJobPayment creates a pre-authorization for a job payment
func AuthorizeJobPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from auth context (you should have auth middleware)
	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.PaymentAuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if paymentService == nil {
		InitPaymentService()
	}

	resp, err := paymentService.AuthorizeJobPayment(userID, req)
	if err != nil {
		log.Printf("Failed to authorize payment: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ==============================================
// PAYMENT CAPTURE (RELEASE FROM ESCROW)
// ==============================================

// CaptureJobPayment captures a previously authorized payment
func CaptureJobPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.PaymentCaptureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if paymentService == nil {
		InitPaymentService()
	}

	resp, err := paymentService.CaptureJobPayment(userID, req)
	if err != nil {
		log.Printf("Failed to capture payment: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ==============================================
// PAYMENT REFUND
// ==============================================

// RefundJobPayment refunds a payment
func RefundJobPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.PaymentRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if paymentService == nil {
		InitPaymentService()
	}

	resp, err := paymentService.RefundJobPayment(userID, req)
	if err != nil {
		log.Printf("Failed to refund payment: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ==============================================
// PAYMENT SUMMARY FOR JOB
// ==============================================

// GetJobPaymentSummary returns payment summary for a job
func GetJobPaymentSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idParam := chi.URLParam(r, "id")
	jobID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid job ID format", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query payment summary using database function
	var summary model.JobPaymentSummary
	query := `SELECT * FROM get_job_payment_summary($1)`
	err = config.DB.QueryRow(query, jobID).Scan(
		&summary.TotalAuthorized,
		&summary.TotalCaptured,
		&summary.TotalRefunded,
		&summary.PlatformFees,
		&summary.WorkerPayment,
		&summary.EscrowStatus,
	)

	if err != nil {
		log.Printf("Failed to get payment summary: %v", err)
		http.Error(w, "Failed to get payment summary", http.StatusInternalServerError)
		return
	}

	summary.JobID = jobID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

// ==============================================
// GET TRANSACTIONS FOR JOB
// ==============================================

// GetJobTransactions returns all transactions for a job
func GetJobTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idParam := chi.URLParam(r, "id")
	jobID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid job ID format", http.StatusBadRequest)
		return
	}

	userID := getUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query transactions
	query := `
		SELECT
			id, uuid, job_id, consumer_id, gig_worker_id, amount, currency,
			status, transaction_type,
			COALESCE(clover_charge_id, '') as clover_charge_id,
			COALESCE(clover_payment_id, '') as clover_payment_id,
			authorized_at, captured_at, capture_amount,
			processing_fee, platform_fee, net_amount,
			escrow_held_at, escrow_released_at,
			refunded_at, refund_amount,
			created_at, updated_at
		FROM transactions
		WHERE job_id = $1
		ORDER BY created_at DESC
	`

	rows, err := config.DB.Query(query, jobID)
	if err != nil {
		log.Printf("Failed to query transactions: %v", err)
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []model.EnhancedTransaction
	for rows.Next() {
		var t model.EnhancedTransaction
		var cloverChargeID, cloverPaymentID string

		err := rows.Scan(
			&t.ID, &t.UUID, &t.JobID, &t.ConsumerID, &t.GigWorkerID,
			&t.Amount, &t.Currency, &t.Status, &t.TransactionType,
			&cloverChargeID, &cloverPaymentID,
			&t.AuthorizedAt, &t.CapturedAt, &t.CaptureAmount,
			&t.ProcessingFee, &t.PlatformFee, &t.NetAmount,
			&t.EscrowHeldAt, &t.EscrowReleasedAt,
			&t.RefundedAt, &t.RefundAmount,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan transaction: %v", err)
			continue
		}

		if cloverChargeID != "" {
			t.CloverChargeID = &cloverChargeID
		}
		if cloverPaymentID != "" {
			t.CloverPaymentID = &cloverPaymentID
		}

		transactions = append(transactions, t)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":       jobID,
		"transactions": transactions,
	})
}

// ==============================================
// HELPER FUNCTIONS
// ==============================================

// getUserIDFromContext gets the user ID from request context
// This is a placeholder - you should implement proper JWT authentication
func getUserIDFromContext(r *http.Request) int {
	// TODO: Implement proper JWT authentication middleware
	// For now, check for a header or query parameter
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}

	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err == nil {
			return userID
		}
	}

	return 0
}
