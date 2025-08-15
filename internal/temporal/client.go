package temporal

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	"app/internal/temporal/workflows"
)

// Client wraps the Temporal client with convenience methods
type Client struct {
	client.Client
}

// NewClient creates a new Temporal client
func NewClient() (*Client, error) {
	temporalHost := getEnv("TEMPORAL_HOST", "localhost:7233")
	
	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	log.Printf("Connected to Temporal server at %s", temporalHost)
	
	return &Client{Client: c}, nil
}

// StartJobWorkflow starts the job lifecycle workflow
func (c *Client) StartJobWorkflow(ctx context.Context, jobID, consumerID int) (client.WorkflowRun, error) {
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("job-%d", jobID),
		TaskQueue: "gigco-jobs",
	}

	we, err := c.ExecuteWorkflow(
		ctx,
		workflowOptions,
		workflows.JobLifecycleWorkflow,
		workflows.JobWorkflowInput{
			JobID:      jobID,
			ConsumerID: consumerID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start job workflow: %w", err)
	}

	log.Printf("Started job workflow for job %d with ID: %s", jobID, we.GetID())
	return we, nil
}

// SignalJobOfferResponse signals the workflow with customer's offer response
func (c *Client) SignalJobOfferResponse(ctx context.Context, workflowID string, accepted bool) error {
	err := c.SignalWorkflow(
		ctx,
		workflowID,
		"",
		"offer-response",
		workflows.OfferResponse{Accepted: accepted},
	)
	if err != nil {
		return fmt.Errorf("failed to signal offer response: %w", err)
	}

	log.Printf("Signaled offer response for workflow %s: accepted=%t", workflowID, accepted)
	return nil
}

// SignalJobStarted signals that a job has started
func (c *Client) SignalJobStarted(ctx context.Context, workflowID string) error {
	err := c.SignalWorkflow(
		ctx,
		workflowID,
		"",
		"job-started",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to signal job started: %w", err)
	}

	log.Printf("Signaled job started for workflow %s", workflowID)
	return nil
}

// SignalJobCompleted signals that a job has been completed
func (c *Client) SignalJobCompleted(ctx context.Context, workflowID string) error {
	err := c.SignalWorkflow(
		ctx,
		workflowID,
		"",
		"job-completed",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to signal job completed: %w", err)
	}

	log.Printf("Signaled job completed for workflow %s", workflowID)
	return nil
}

// SignalReviewSubmitted signals that a review has been submitted
func (c *Client) SignalReviewSubmitted(ctx context.Context, workflowID string, review workflows.ReviewSubmission) error {
	err := c.SignalWorkflow(
		ctx,
		workflowID,
		"",
		"review-submitted",
		review,
	)
	if err != nil {
		return fmt.Errorf("failed to signal review submitted: %w", err)
	}

	log.Printf("Signaled review submitted for workflow %s", workflowID)
	return nil
}

// GetWorkflowStatus retrieves the workflow status
func (c *Client) GetWorkflowStatus(ctx context.Context, workflowID string) error {
	// This is a utility method for debugging workflows
	workflowRun := c.GetWorkflow(ctx, workflowID, "")
	
	var result interface{}
	err := workflowRun.Get(ctx, &result)
	if err != nil {
		log.Printf("Workflow %s is still running or failed: %v", workflowID, err)
		return err
	}
	
	log.Printf("Workflow %s completed with result: %v", workflowID, result)
	return nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}