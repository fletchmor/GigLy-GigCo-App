package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"app/config"
	"app/internal/model"
)

// CloverService handles all Clover API interactions
type CloverService struct {
	config     *config.CloverConfig
	httpClient *http.Client
}

// NewCloverService creates a new Clover service instance
func NewCloverService(cfg *config.CloverConfig) *CloverService {
	return &CloverService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ==============================================
// TOKENIZATION
// ==============================================

// TokenizeCard tokenizes a credit card and returns a Clover token
func (s *CloverService) TokenizeCard(card model.CloverCard) (*model.CloverTokenizeResponse, error) {
	reqBody := model.CloverTokenizeRequest{
		Card: card,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tokenize request: %w", err)
	}

	req, err := http.NewRequest("POST", s.config.TokenizationEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenize request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.config.APIAccessKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tokenize request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tokenize response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tokenization failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var tokenResp model.CloverTokenizeResponse
	if err := json.Unmarshal(responseBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokenize response: %w", err)
	}

	return &tokenResp, nil
}

// ==============================================
// AUTHORIZATION (PRE-AUTH)
// ==============================================

// AuthorizePayment creates a pre-authorization (hold) on a card
func (s *CloverService) AuthorizePayment(token string, amountCents int64, metadata map[string]interface{}) (*model.CloverChargeResponse, error) {
	reqBody := model.CloverChargeRequest{
		Amount:   amountCents,
		Currency: "USD",
		Source:   token,
		Capture:  false, // false for pre-authorization
		Metadata: metadata,
	}

	return s.createCharge(reqBody)
}

// ==============================================
// DIRECT CHARGE
// ==============================================

// ChargePayment creates a direct charge (authorization + capture)
func (s *CloverService) ChargePayment(token string, amountCents int64, metadata map[string]interface{}) (*model.CloverChargeResponse, error) {
	reqBody := model.CloverChargeRequest{
		Amount:   amountCents,
		Currency: "USD",
		Source:   token,
		Capture:  true, // true for direct charge
		Metadata: metadata,
	}

	return s.createCharge(reqBody)
}

// createCharge is a helper method to create a charge (used by both authorize and direct charge)
func (s *CloverService) createCharge(reqBody model.CloverChargeRequest) (*model.CloverChargeResponse, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal charge request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/charges", s.config.APIEndpoint)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create charge request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.AccessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute charge request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read charge response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("charge failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var chargeResp model.CloverChargeResponse
	if err := json.Unmarshal(responseBody, &chargeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal charge response: %w", err)
	}

	return &chargeResp, nil
}

// ==============================================
// CAPTURE
// ==============================================

// CapturePayment captures a previously authorized payment
func (s *CloverService) CapturePayment(paymentID string, amountCents *int64) (*model.CloverCaptureResponse, error) {
	var reqBody model.CloverCaptureRequest
	if amountCents != nil {
		reqBody.Amount = *amountCents
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal capture request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/payments/%s/capture", s.config.APIEndpoint, paymentID)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create capture request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.AccessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute capture request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read capture response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("capture failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var captureResp model.CloverCaptureResponse
	if err := json.Unmarshal(responseBody, &captureResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capture response: %w", err)
	}

	return &captureResp, nil
}

// ==============================================
// REFUNDS
// ==============================================

// RefundPayment refunds a charge
func (s *CloverService) RefundPayment(chargeID string, amountCents *int64, reason string) (*model.CloverRefundResponse, error) {
	reqBody := model.CloverRefundRequest{
		ChargeID: chargeID,
		Reason:   reason,
	}

	if amountCents != nil {
		reqBody.Amount = *amountCents
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refund request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/refunds", s.config.APIEndpoint)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create refund request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.AccessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute refund request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refund response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("refund failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var refundResp model.CloverRefundResponse
	if err := json.Unmarshal(responseBody, &refundResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refund response: %w", err)
	}

	return &refundResp, nil
}

// ==============================================
// HELPER FUNCTIONS
// ==============================================

// DollarsToCents converts dollars to cents for Clover API
func DollarsToCents(dollars float64) int64 {
	return int64(dollars * 100)
}

// CentsToDollars converts cents to dollars from Clover API
func CentsToDollars(cents int64) float64 {
	return float64(cents) / 100.0
}
