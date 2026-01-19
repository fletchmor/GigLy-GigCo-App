package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// PushService handles push notifications via Firebase Cloud Messaging
type PushService struct {
	serverKey  string
	projectID  string
	httpClient *http.Client
	fcmURL     string
}

// PushConfig holds push notification configuration
type PushConfig struct {
	ServerKey string // FCM Server Key (legacy) or Service Account Key
	ProjectID string // Firebase Project ID
}

// NewPushService creates a new push notification service
func NewPushService(cfg PushConfig) (*PushService, error) {
	if cfg.ServerKey == "" {
		return nil, fmt.Errorf("FCM server key is required")
	}

	return &PushService{
		serverKey:  cfg.ServerKey,
		projectID:  cfg.ProjectID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		fcmURL:     "https://fcm.googleapis.com/fcm/send",
	}, nil
}

// NewPushServiceFromEnv creates push service from environment variables
func NewPushServiceFromEnv() (*PushService, error) {
	return NewPushService(PushConfig{
		ServerKey: os.Getenv("FCM_SERVER_KEY"),
		ProjectID: os.Getenv("FIREBASE_PROJECT_ID"),
	})
}

// FCMMessage represents a Firebase Cloud Messaging message
type FCMMessage struct {
	To               string            `json:"to,omitempty"`
	RegistrationIDs  []string          `json:"registration_ids,omitempty"`
	Condition        string            `json:"condition,omitempty"`
	Topic            string            `json:"topic,omitempty"`
	Notification     *FCMNotification  `json:"notification,omitempty"`
	Data             map[string]string `json:"data,omitempty"`
	Priority         string            `json:"priority,omitempty"`
	TimeToLive       int               `json:"time_to_live,omitempty"`
	ContentAvailable bool              `json:"content_available,omitempty"`
	MutableContent   bool              `json:"mutable_content,omitempty"`
}

// FCMNotification represents the notification payload
type FCMNotification struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Badge        string `json:"badge,omitempty"`
	ClickAction  string `json:"click_action,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs string `json:"title_loc_args,omitempty"`
}

// FCMResponse represents the response from FCM
type FCMResponse struct {
	MulticastID  int64        `json:"multicast_id"`
	Success      int          `json:"success"`
	Failure      int          `json:"failure"`
	CanonicalIDs int          `json:"canonical_ids"`
	Results      []FCMResult  `json:"results"`
}

// FCMResult represents individual result for each recipient
type FCMResult struct {
	MessageID      string `json:"message_id,omitempty"`
	RegistrationID string `json:"registration_id,omitempty"`
	Error          string `json:"error,omitempty"`
}

// SendToDevice sends a push notification to a specific device
func (s *PushService) SendToDevice(deviceToken string, notification *FCMNotification, data map[string]string) (*FCMResponse, error) {
	message := FCMMessage{
		To:           deviceToken,
		Notification: notification,
		Data:         data,
		Priority:     "high",
	}

	return s.send(message)
}

// SendToDevices sends a push notification to multiple devices
func (s *PushService) SendToDevices(deviceTokens []string, notification *FCMNotification, data map[string]string) (*FCMResponse, error) {
	if len(deviceTokens) > 1000 {
		return nil, fmt.Errorf("cannot send to more than 1000 devices at once")
	}

	message := FCMMessage{
		RegistrationIDs: deviceTokens,
		Notification:    notification,
		Data:            data,
		Priority:        "high",
	}

	return s.send(message)
}

// SendToTopic sends a push notification to a topic
func (s *PushService) SendToTopic(topic string, notification *FCMNotification, data map[string]string) (*FCMResponse, error) {
	message := FCMMessage{
		To:           "/topics/" + topic,
		Notification: notification,
		Data:         data,
		Priority:     "high",
	}

	return s.send(message)
}

// send sends the FCM message
func (s *PushService) send(message FCMMessage) (*FCMResponse, error) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", s.fcmURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "key="+s.serverKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FCM returned status %d", resp.StatusCode)
	}

	var fcmResp FCMResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &fcmResp, nil
}

// JobNotification creates a notification for job-related events
type JobNotification struct {
	JobID       string
	JobTitle    string
	Message     string
	ActionType  string // "view", "accept", "complete", etc.
}

// SendJobNotification sends a job-related push notification
func (s *PushService) SendJobNotification(deviceToken string, jn JobNotification) (*FCMResponse, error) {
	notification := &FCMNotification{
		Title: "GigCo: " + jn.JobTitle,
		Body:  jn.Message,
		Sound: "default",
	}

	data := map[string]string{
		"job_id":      jn.JobID,
		"job_title":   jn.JobTitle,
		"action_type": jn.ActionType,
		"type":        "job_notification",
	}

	return s.SendToDevice(deviceToken, notification, data)
}

// PaymentNotification creates a notification for payment events
type PaymentNotification struct {
	TransactionID string
	Amount        string
	Message       string
}

// SendPaymentNotification sends a payment-related push notification
func (s *PushService) SendPaymentNotification(deviceToken string, pn PaymentNotification) (*FCMResponse, error) {
	notification := &FCMNotification{
		Title: "GigCo Payment",
		Body:  pn.Message,
		Sound: "default",
	}

	data := map[string]string{
		"transaction_id": pn.TransactionID,
		"amount":         pn.Amount,
		"type":           "payment_notification",
	}

	return s.SendToDevice(deviceToken, notification, data)
}

// MockPushService is a mock push service for testing
type MockPushService struct {
	SentNotifications []FCMMessage
}

// SendToDevice mocks sending a notification
func (m *MockPushService) SendToDevice(deviceToken string, notification *FCMNotification, data map[string]string) (*FCMResponse, error) {
	m.SentNotifications = append(m.SentNotifications, FCMMessage{
		To:           deviceToken,
		Notification: notification,
		Data:         data,
	})
	return &FCMResponse{Success: 1}, nil
}
