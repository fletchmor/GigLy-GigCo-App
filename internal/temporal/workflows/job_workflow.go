package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// JobWorkflowInput contains the input for a job workflow
type JobWorkflowInput struct {
	JobID      int `json:"job_id"`
	ConsumerID int `json:"consumer_id"`
}

// JobWorkflowState tracks the current state of the job
type JobWorkflowState struct {
	JobID            int     `json:"job_id"`
	CurrentState     string  `json:"current_state"`
	PricedAmount     float64 `json:"priced_amount"`
	AssignedWorkerID int     `json:"assigned_worker_id"`
	PaymentID        string  `json:"payment_id"`
	ReviewsReceived  int     `json:"reviews_received"`
}

// PriceJobResult contains the result of pricing a job
type PriceJobResult struct {
	JobID  int     `json:"job_id"`
	Amount float64 `json:"amount"`
}

// MatchWorkerResult contains the result of worker matching
type MatchWorkerResult struct {
	JobID    int `json:"job_id"`
	WorkerID int `json:"worker_id"`
}

// ProcessPaymentResult contains the result of payment processing
type ProcessPaymentResult struct {
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
}

// OfferResponse represents customer response to job offer
type OfferResponse struct {
	Accepted bool `json:"accepted"`
}

// ReviewSubmission represents a review submission
type ReviewSubmission struct {
	JobID      int    `json:"job_id"`
	ReviewerID int    `json:"reviewer_id"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
}

// JobLifecycleWorkflow orchestrates the entire job lifecycle
func JobLifecycleWorkflow(ctx workflow.Context, input JobWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting job workflow", "jobID", input.JobID)

	// Set workflow options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	state := &JobWorkflowState{
		JobID:        input.JobID,
		CurrentState: "draft",
	}

	// Step 1: Price the job
	var priceResult PriceJobResult
	err := workflow.ExecuteActivity(ctx, "PriceJob", input.JobID).Get(ctx, &priceResult)
	if err != nil {
		logger.Error("Failed to price job", "error", err)
		return err
	}
	state.PricedAmount = priceResult.Amount
	state.CurrentState = "priced"
	logger.Info("Job priced", "jobID", input.JobID, "amount", priceResult.Amount)

	// Step 2: Send offer to customer and wait for response
	err = workflow.ExecuteActivity(ctx, "SendJobOffer", input.JobID, priceResult.Amount).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to send job offer", "error", err)
		return err
	}

	// Wait for customer decision (with timeout)
	selector := workflow.NewSelector(ctx)
	var offerAccepted bool

	offerChannel := workflow.GetSignalChannel(ctx, "offer-response")
	selector.AddReceive(offerChannel, func(c workflow.ReceiveChannel, more bool) {
		var response OfferResponse
		c.Receive(ctx, &response)
		offerAccepted = response.Accepted
	})

	// Add timeout for offer response (24 hours)
	timerFuture := workflow.NewTimer(ctx, 24*time.Hour)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		offerAccepted = false
		logger.Info("Offer timeout reached", "jobID", input.JobID)
	})

	selector.Select(ctx)

	if !offerAccepted {
		state.CurrentState = "rejected"
		logger.Info("Job offer rejected or timed out", "jobID", input.JobID)
		return workflow.ExecuteActivity(ctx, "HandleJobRejection", input.JobID).Get(ctx, nil)
	}

	state.CurrentState = "accepted"
	logger.Info("Job offer accepted", "jobID", input.JobID)

	// Step 3: Find and assign worker
	retryCount := 0
	maxRetries := 5

	for retryCount < maxRetries {
		var matchResult MatchWorkerResult
		err = workflow.ExecuteActivity(ctx, "FindMatchingWorker", input.JobID).Get(ctx, &matchResult)

		if err == nil && matchResult.WorkerID > 0 {
			state.AssignedWorkerID = matchResult.WorkerID
			state.CurrentState = "worker_assigned"
			logger.Info("Worker assigned", "jobID", input.JobID, "workerID", matchResult.WorkerID)
			break
		}

		// Wait before retry with exponential backoff
		retryDelay := time.Duration(retryCount+1) * 5 * time.Minute
		workflow.Sleep(ctx, retryDelay)
		retryCount++
		logger.Info("Retrying worker assignment", "jobID", input.JobID, "attempt", retryCount)
	}

	if state.AssignedWorkerID == 0 {
		logger.Error("No worker found after retries", "jobID", input.JobID)
		state.CurrentState = "no_worker_available"
		return workflow.ExecuteActivity(ctx, "HandleNoWorkerAvailable", input.JobID).Get(ctx, nil)
	}

	// Step 4: Schedule the job
	err = workflow.ExecuteActivity(ctx, "ScheduleJob", input.JobID, state.AssignedWorkerID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to schedule job", "error", err)
		return err
	}
	state.CurrentState = "scheduled"
	logger.Info("Job scheduled", "jobID", input.JobID)

	// Step 5: Wait for job to start
	startSignal := workflow.GetSignalChannel(ctx, "job-started")
	startSignal.Receive(ctx, nil)
	state.CurrentState = "in_progress"
	logger.Info("Job started", "jobID", input.JobID)

	// Step 6: Wait for job completion
	completionSignal := workflow.GetSignalChannel(ctx, "job-completed")
	completionSignal.Receive(ctx, nil)
	state.CurrentState = "completed"
	logger.Info("Job completed", "jobID", input.JobID)

	// Step 7: Process payment
	var paymentResult ProcessPaymentResult
	err = workflow.ExecuteActivity(ctx, "ProcessJobPayment", input.JobID).Get(ctx, &paymentResult)
	if err != nil {
		logger.Error("Payment failed", "error", err)
		// Continue workflow to handle payment retry separately
		return workflow.ExecuteActivity(ctx, "HandlePaymentFailure", input.JobID).Get(ctx, nil)
	}
	state.PaymentID = paymentResult.TransactionID
	state.CurrentState = "paid"
	logger.Info("Job payment processed", "jobID", input.JobID, "transactionID", paymentResult.TransactionID)

	// Step 8: Request reviews
	err = workflow.ExecuteActivity(ctx, "RequestReviews", input.JobID).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to request reviews", "error", err)
		// Continue workflow even if review request fails
	}
	state.CurrentState = "review_pending"

	// Step 9: Wait for reviews (with timeout)
	reviewTimer := workflow.NewTimer(ctx, 7*24*time.Hour) // 7 days
	reviewChannel := workflow.GetSignalChannel(ctx, "review-submitted")

	reviewsReceived := 0
	maxReviews := 2 // Both consumer and worker reviews

	for reviewsReceived < maxReviews {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(reviewChannel, func(c workflow.ReceiveChannel, more bool) {
			var review ReviewSubmission
			c.Receive(ctx, &review)
			reviewsReceived++
			state.ReviewsReceived = reviewsReceived
			logger.Info("Review received", "jobID", input.JobID, "reviewsReceived", reviewsReceived)
		})

		selector.AddFuture(reviewTimer, func(f workflow.Future) {
			// Timeout reached, close without all reviews
			logger.Info("Review timeout reached", "jobID", input.JobID, "reviewsReceived", reviewsReceived)
			reviewsReceived = maxReviews
		})

		selector.Select(ctx)
	}

	// Step 10: Close the job
	err = workflow.ExecuteActivity(ctx, "CloseJob", input.JobID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to close job", "error", err)
		return err
	}
	state.CurrentState = "closed"

	logger.Info("Job workflow completed successfully", "jobID", input.JobID, "finalState", state.CurrentState)
	return nil
}

// PaymentRetryWorkflow handles payment retry logic
func PaymentRetryWorkflow(ctx workflow.Context, input JobWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting payment retry workflow", "jobID", input.JobID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 5,
			InitialInterval: time.Minute,
			BackoffCoefficient: 2.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var paymentResult ProcessPaymentResult
	err := workflow.ExecuteActivity(ctx, "ProcessJobPayment", input.JobID).Get(ctx, &paymentResult)
	if err != nil {
		logger.Error("Payment retry failed", "jobID", input.JobID, "error", err)
		return workflow.ExecuteActivity(ctx, "HandlePaymentFailure", input.JobID).Get(ctx, nil)
	}

	logger.Info("Payment retry successful", "jobID", input.JobID, "transactionID", paymentResult.TransactionID)
	return workflow.ExecuteActivity(ctx, "UpdateJobPaymentStatus", input.JobID, paymentResult.TransactionID).Get(ctx, nil)
}