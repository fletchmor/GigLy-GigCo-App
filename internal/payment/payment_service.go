package payment

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"app/config"
	"app/internal/model"
)

// PaymentService handles payment business logic and database operations
type PaymentService struct {
	db            *sql.DB
	cloverService *CloverService
	config        *config.CloverConfig
}

// NewPaymentService creates a new payment service instance
func NewPaymentService(db *sql.DB, cloverConfig *config.CloverConfig) *PaymentService {
	return &PaymentService{
		db:            db,
		cloverService: NewCloverService(cloverConfig),
		config:        cloverConfig,
	}
}

// ==============================================
// AUTHORIZATION (ESCROW)
// ==============================================

// AuthorizeJobPayment creates a pre-authorization for a job payment
func (s *PaymentService) AuthorizeJobPayment(userID int, req model.PaymentAuthorizeRequest) (*model.PaymentAuthorizeResponse, error) {
	// 1. Get job details
	job, err := s.getJob(req.JobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Verify user is the consumer
	if job.ConsumerID != userID {
		return nil, fmt.Errorf("unauthorized: user is not the consumer of this job")
	}

	// 2. Get or create card token
	var cardToken string
	if req.CardToken != nil {
		cardToken = *req.CardToken
	} else if req.CardDetails != nil {
		tokenResp, err := s.cloverService.TokenizeCard(model.CloverCard{
			Number:   req.CardDetails.Number,
			ExpMonth: req.CardDetails.ExpMonth,
			ExpYear:  req.CardDetails.ExpYear,
			CVV:      req.CardDetails.CVV,
			Name:     req.CardDetails.Name,
			AddressLine1: req.CardDetails.AddressLine1,
			AddressCity:  req.CardDetails.AddressCity,
			AddressState: req.CardDetails.AddressState,
			AddressZip:   req.CardDetails.AddressZip,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to tokenize card: %w", err)
		}
		cardToken = tokenResp.ID

		// Save card if requested
		if req.SaveCard {
			if err := s.savePaymentMethod(userID, tokenResp, req.PaymentMethodID); err != nil {
				// Log error but don't fail the transaction
				fmt.Printf("Warning: failed to save payment method: %v\n", err)
			}
		}
	} else if req.PaymentMethodID != nil {
		// Load saved payment method
		pm, err := s.getPaymentMethod(*req.PaymentMethodID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment method: %w", err)
		}
		if pm.CloverToken != nil {
			cardToken = *pm.CloverToken
		} else {
			return nil, fmt.Errorf("payment method does not have a valid token")
		}
	} else {
		return nil, fmt.Errorf("no payment source provided")
	}

	// 3. Calculate fees
	netAmount, platformFee, processingFee := s.config.CalculateNetAmount(req.Amount)

	// 4. Create Clover authorization
	metadata := map[string]interface{}{
		"job_id":      req.JobID,
		"consumer_id": userID,
		"type":        "job_payment",
	}
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	cloverResp, err := s.cloverService.AuthorizePayment(
		cardToken,
		DollarsToCents(req.Amount),
		metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize payment with Clover: %w", err)
	}

	// 5. Create transaction record
	now := time.Now()
	authExpiresAt := now.Add(7 * 24 * time.Hour) // Typical 7-day auth window

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var transactionID int
	err = tx.QueryRow(`
		INSERT INTO transactions (
			job_id, consumer_id, gig_worker_id, amount, currency,
			status, transaction_type,
			clover_charge_id, clover_source_token,
			authorized_at, authorization_expires_at,
			payment_method, last_four,
			processing_fee, platform_fee, net_amount,
			escrow_held_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id
	`,
		req.JobID, job.ConsumerID, job.GigWorkerID, req.Amount, "USD",
		"completed", "authorization",
		cloverResp.ID, cloverResp.Source.ID,
		now, authExpiresAt,
		cloverResp.Source.Brand, cloverResp.Source.Last4,
		processingFee, platformFee, netAmount,
		now, toJSON(metadata),
	).Scan(&transactionID)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// 6. Create payment event log
	if err := s.createPaymentEvent(tx, transactionID, "authorize", "success", cloverResp, nil, userID); err != nil {
		return nil, fmt.Errorf("failed to create payment event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 7. Get full transaction details
	transaction, err := s.getTransaction(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &model.PaymentAuthorizeResponse{
		Success:       true,
		TransactionID: transactionID,
		Transaction:   transaction,
		Message:       "Payment authorized successfully. Funds are held in escrow.",
	}, nil
}

// ==============================================
// CAPTURE (RELEASE FROM ESCROW)
// ==============================================

// CaptureJobPayment captures a previously authorized payment
func (s *PaymentService) CaptureJobPayment(userID int, req model.PaymentCaptureRequest) (*model.PaymentCaptureResponse, error) {
	// 1. Get transaction
	transaction, err := s.getTransaction(req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Verify transaction type
	if transaction.TransactionType != model.TransactionTypeAuthorization {
		return nil, fmt.Errorf("transaction is not an authorization")
	}

	// Verify not already captured
	if transaction.CapturedAt != nil {
		return nil, fmt.Errorf("transaction already captured")
	}

	// 2. Get job and verify permissions
	job, err := s.getJob(transaction.JobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Either consumer or worker can trigger capture (when job is completed)
	isConsumer := job.ConsumerID == userID
	isWorker := job.GigWorkerID != nil && *job.GigWorkerID == userID
	if !isConsumer && !isWorker {
		return nil, fmt.Errorf("unauthorized: user cannot capture this payment")
	}

	// 3. Determine capture amount
	var captureAmountCents *int64
	if req.Amount != nil {
		cents := DollarsToCents(*req.Amount)
		captureAmountCents = &cents
	}

	// 4. Capture with Clover
	if transaction.CloverPaymentID == nil {
		return nil, fmt.Errorf("transaction does not have a Clover payment ID")
	}

	cloverResp, err := s.cloverService.CapturePayment(*transaction.CloverPaymentID, captureAmountCents)
	if err != nil {
		// Log the failure
		s.createPaymentEventSimple(req.TransactionID, "capture", "failed", nil, err, userID)
		return nil, fmt.Errorf("failed to capture payment with Clover: %w", err)
	}

	// 5. Update transaction
	now := time.Now()
	captureAmount := CentsToDollars(cloverResp.Amount)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE transactions
		SET captured_at = $1, capture_amount = $2, escrow_released_at = $3, updated_at = $4
		WHERE id = $5
	`, now, captureAmount, now, now, req.TransactionID)

	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// 6. Create capture event log
	if err := s.createPaymentEvent(tx, req.TransactionID, "capture", "success", cloverResp, nil, userID); err != nil {
		return nil, fmt.Errorf("failed to create payment event: %w", err)
	}

	// 7. Update job status to paid
	_, err = tx.Exec(`UPDATE jobs SET status = 'paid', updated_at = $1 WHERE id = $2`, now, job.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 8. Get updated transaction
	updatedTransaction, err := s.getTransaction(req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated transaction: %w", err)
	}

	return &model.PaymentCaptureResponse{
		Success:       true,
		TransactionID: req.TransactionID,
		Transaction:   updatedTransaction,
		Message:       "Payment captured successfully. Funds released from escrow.",
	}, nil
}

// ==============================================
// REFUNDS
// ==============================================

// RefundJobPayment refunds a payment
func (s *PaymentService) RefundJobPayment(userID int, req model.PaymentRefundRequest) (*model.PaymentRefundResponse, error) {
	// 1. Get transaction
	transaction, err := s.getTransaction(req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// 2. Get job and verify permissions (typically only consumer can refund)
	job, err := s.getJob(transaction.JobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if job.ConsumerID != userID {
		return nil, fmt.Errorf("unauthorized: only the consumer can request a refund")
	}

	// 3. Verify can be refunded
	if transaction.Status == model.TransactionStatusRefunded {
		return nil, fmt.Errorf("transaction already refunded")
	}

	if transaction.CloverChargeID == nil {
		return nil, fmt.Errorf("transaction does not have a Clover charge ID")
	}

	// 4. Determine refund amount
	var refundAmountCents *int64
	if req.Amount != nil {
		cents := DollarsToCents(*req.Amount)
		refundAmountCents = &cents
	}

	// 5. Process refund with Clover
	cloverResp, err := s.cloverService.RefundPayment(*transaction.CloverChargeID, refundAmountCents, req.Reason)
	if err != nil {
		s.createPaymentEventSimple(req.TransactionID, "refund", "failed", nil, err, userID)
		return nil, fmt.Errorf("failed to refund payment with Clover: %w", err)
	}

	// 6. Create refund transaction
	now := time.Now()
	refundAmount := CentsToDollars(cloverResp.Amount)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var refundID int
	err = tx.QueryRow(`
		INSERT INTO transactions (
			job_id, consumer_id, gig_worker_id, amount, currency,
			status, transaction_type,
			clover_refund_id,
			refunded_at, refund_amount, refund_reason,
			parent_transaction_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`,
		job.ID, job.ConsumerID, job.GigWorkerID, refundAmount, "USD",
		"completed", "refund",
		cloverResp.ID,
		now, refundAmount, req.Reason,
		req.TransactionID,
	).Scan(&refundID)

	if err != nil {
		return nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}

	// 7. Update original transaction status
	_, err = tx.Exec(`
		UPDATE transactions
		SET status = 'refunded', refunded_at = $1, refund_amount = $2, refund_reason = $3, updated_at = $4
		WHERE id = $5
	`, now, refundAmount, req.Reason, now, req.TransactionID)

	if err != nil {
		return nil, fmt.Errorf("failed to update original transaction: %w", err)
	}

	// 8. Create refund event log
	if err := s.createPaymentEvent(tx, refundID, "refund", "success", cloverResp, nil, userID); err != nil {
		return nil, fmt.Errorf("failed to create payment event: %w", err)
	}

	// 9. Update job status
	_, err = tx.Exec(`UPDATE jobs SET status = 'cancelled', updated_at = $1 WHERE id = $2`, now, job.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 10. Get refund transaction
	refundTransaction, err := s.getTransaction(refundID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refund transaction: %w", err)
	}

	return &model.PaymentRefundResponse{
		Success:     true,
		RefundID:    refundID,
		Transaction: refundTransaction,
		Message:     "Payment refunded successfully.",
	}, nil
}

// ==============================================
// HELPER METHODS
// ==============================================

func (s *PaymentService) getJob(jobID int) (*model.Job, error) {
	var job model.Job
	err := s.db.QueryRow(`
		SELECT id, uuid, consumer_id, gig_worker_id, title, description, status
		FROM jobs WHERE id = $1
	`, jobID).Scan(
		&job.ID, &job.UUID, &job.ConsumerID, &job.GigWorkerID,
		&job.Title, &job.Description, &job.Status,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *PaymentService) getTransaction(id int) (*model.EnhancedTransaction, error) {
	var t model.EnhancedTransaction
	err := s.db.QueryRow(`
		SELECT id, uuid, job_id, consumer_id, gig_worker_id, amount, currency,
		       status, transaction_type, clover_charge_id, clover_payment_id,
		       authorized_at, captured_at, capture_amount,
		       processing_fee, platform_fee, net_amount,
		       escrow_held_at, escrow_released_at,
		       created_at, updated_at
		FROM transactions WHERE id = $1
	`, id).Scan(
		&t.ID, &t.UUID, &t.JobID, &t.ConsumerID, &t.GigWorkerID, &t.Amount, &t.Currency,
		&t.Status, &t.TransactionType, &t.CloverChargeID, &t.CloverPaymentID,
		&t.AuthorizedAt, &t.CapturedAt, &t.CaptureAmount,
		&t.ProcessingFee, &t.PlatformFee, &t.NetAmount,
		&t.EscrowHeldAt, &t.EscrowReleasedAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
	return &t, err
}

func (s *PaymentService) getPaymentMethod(id, userID int) (*model.UserPaymentMethod, error) {
	var pm model.UserPaymentMethod
	err := s.db.QueryRow(`
		SELECT id, uuid, user_id, clover_token, type, last_four, brand, is_default, is_active
		FROM user_payment_methods
		WHERE id = $1 AND user_id = $2 AND is_active = true
	`, id, userID).Scan(
		&pm.ID, &pm.UUID, &pm.UserID, &pm.CloverToken,
		&pm.Type, &pm.LastFour, &pm.Brand, &pm.IsDefault, &pm.IsActive,
	)
	return &pm, err
}

func (s *PaymentService) savePaymentMethod(userID int, tokenResp *model.CloverTokenizeResponse, existingID *int) error {
	// Implementation for saving payment method
	// This would insert/update user_payment_methods table
	return nil
}

func (s *PaymentService) createPaymentEvent(tx *sql.Tx, transactionID int, eventType, status string, response interface{}, err error, userID int) error {
	var errorMsg *string
	if err != nil {
		msg := err.Error()
		errorMsg = &msg
	}

	_, execErr := tx.Exec(`
		INSERT INTO payment_events (transaction_id, event_type, event_status, clover_response, error_message, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, transactionID, eventType, status, toJSON(response), errorMsg, userID)

	return execErr
}

func (s *PaymentService) createPaymentEventSimple(transactionID int, eventType, status string, response interface{}, err error, userID int) {
	var errorMsg *string
	if err != nil {
		msg := err.Error()
		errorMsg = &msg
	}

	s.db.Exec(`
		INSERT INTO payment_events (transaction_id, event_type, event_status, clover_response, error_message, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, transactionID, eventType, status, toJSON(response), errorMsg, userID)
}

func toJSON(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	data, _ := json.Marshal(v)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}
