package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

// Service handles email sending operations
type Service struct {
	apiKey     string
	fromEmail  string
	fromName   string
	baseURL    string
	httpClient *http.Client
}

// Config holds email service configuration
type Config struct {
	APIKey    string
	FromEmail string
	FromName  string
	Provider  string // "sendgrid" or "ses" (future)
}

// NewService creates a new email service
func NewService(cfg Config) (*Service, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("email API key is required")
	}
	if cfg.FromEmail == "" {
		return nil, fmt.Errorf("from email is required")
	}

	fromName := cfg.FromName
	if fromName == "" {
		fromName = "GigCo"
	}

	return &Service{
		apiKey:    cfg.APIKey,
		fromEmail: cfg.FromEmail,
		fromName:  fromName,
		baseURL:   "https://api.sendgrid.com/v3",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// NewServiceFromEnv creates email service from environment variables
func NewServiceFromEnv() (*Service, error) {
	return NewService(Config{
		APIKey:    os.Getenv("SENDGRID_API_KEY"),
		FromEmail: os.Getenv("EMAIL_FROM"),
		FromName:  os.Getenv("EMAIL_FROM_NAME"),
		Provider:  "sendgrid",
	})
}

// SendGridRequest represents a SendGrid API request
type SendGridRequest struct {
	Personalizations []Personalization `json:"personalizations"`
	From             EmailAddress      `json:"from"`
	Subject          string            `json:"subject"`
	Content          []Content         `json:"content"`
}

// Personalization represents email recipients
type Personalization struct {
	To []EmailAddress `json:"to"`
}

// EmailAddress represents an email address
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Content represents email content
type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Send sends an email
func (s *Service) Send(to, toName, subject, htmlContent, textContent string) error {
	request := SendGridRequest{
		Personalizations: []Personalization{
			{
				To: []EmailAddress{
					{Email: to, Name: toName},
				},
			},
		},
		From: EmailAddress{
			Email: s.fromEmail,
			Name:  s.fromName,
		},
		Subject: subject,
		Content: []Content{},
	}

	// Add text content if provided
	if textContent != "" {
		request.Content = append(request.Content, Content{
			Type:  "text/plain",
			Value: textContent,
		})
	}

	// Add HTML content
	if htmlContent != "" {
		request.Content = append(request.Content, Content{
			Type:  "text/html",
			Value: htmlContent,
		})
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL+"/mail/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("email API returned status %d", resp.StatusCode)
	}

	return nil
}

// VerificationEmailData holds data for verification email template
type VerificationEmailData struct {
	UserName         string
	VerificationLink string
	ExpirationHours  int
}

// SendVerificationEmail sends an email verification email
func (s *Service) SendVerificationEmail(to, userName, token string) error {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://app.gigco.com"
	}

	data := VerificationEmailData{
		UserName:         userName,
		VerificationLink: fmt.Sprintf("%s/verify-email?token=%s", baseURL, token),
		ExpirationHours:  24,
	}

	htmlContent, err := renderTemplate("verification", data)
	if err != nil {
		// Fallback to simple HTML
		htmlContent = fmt.Sprintf(`
			<h1>Welcome to GigCo, %s!</h1>
			<p>Please verify your email address by clicking the link below:</p>
			<p><a href="%s">Verify Email Address</a></p>
			<p>This link will expire in %d hours.</p>
			<p>If you didn't create an account with GigCo, please ignore this email.</p>
		`, data.UserName, data.VerificationLink, data.ExpirationHours)
	}

	textContent := fmt.Sprintf(
		"Welcome to GigCo, %s!\n\nPlease verify your email by visiting: %s\n\nThis link expires in %d hours.",
		data.UserName, data.VerificationLink, data.ExpirationHours,
	)

	return s.Send(to, userName, "Verify your GigCo email address", htmlContent, textContent)
}

// PasswordResetData holds data for password reset email template
type PasswordResetData struct {
	UserName        string
	ResetLink       string
	ExpirationMins  int
	IPAddress       string
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(to, userName, token, ipAddress string) error {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://app.gigco.com"
	}

	data := PasswordResetData{
		UserName:       userName,
		ResetLink:      fmt.Sprintf("%s/reset-password?token=%s", baseURL, token),
		ExpirationMins: 30,
		IPAddress:      ipAddress,
	}

	htmlContent, err := renderTemplate("password_reset", data)
	if err != nil {
		// Fallback to simple HTML
		htmlContent = fmt.Sprintf(`
			<h1>Password Reset Request</h1>
			<p>Hi %s,</p>
			<p>We received a request to reset your password. Click the link below to set a new password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>This link will expire in %d minutes.</p>
			<p>If you didn't request a password reset, please ignore this email or contact support if you're concerned.</p>
			<p><small>Request originated from IP: %s</small></p>
		`, data.UserName, data.ResetLink, data.ExpirationMins, data.IPAddress)
	}

	textContent := fmt.Sprintf(
		"Hi %s,\n\nWe received a request to reset your password.\n\nReset your password here: %s\n\nThis link expires in %d minutes.\n\nRequest from IP: %s",
		data.UserName, data.ResetLink, data.ExpirationMins, data.IPAddress,
	)

	return s.Send(to, userName, "Reset your GigCo password", htmlContent, textContent)
}

// JobNotificationData holds data for job notification emails
type JobNotificationData struct {
	UserName    string
	JobTitle    string
	JobID       string
	Message     string
	ActionLink  string
	ActionLabel string
}

// SendJobNotification sends a job-related notification email
func (s *Service) SendJobNotification(to, userName string, data JobNotificationData) error {
	htmlContent := fmt.Sprintf(`
		<h1>Job Update</h1>
		<p>Hi %s,</p>
		<p>%s</p>
		<p><strong>Job:</strong> %s</p>
		<p><a href="%s">%s</a></p>
	`, data.UserName, data.Message, data.JobTitle, data.ActionLink, data.ActionLabel)

	textContent := fmt.Sprintf(
		"Hi %s,\n\n%s\n\nJob: %s\n\nView details: %s",
		data.UserName, data.Message, data.JobTitle, data.ActionLink,
	)

	return s.Send(to, userName, fmt.Sprintf("GigCo: %s", data.JobTitle), htmlContent, textContent)
}

// renderTemplate renders an email template
func renderTemplate(name string, data interface{}) (string, error) {
	templatePath := fmt.Sprintf("templates/email/%s.html", name)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// MockService is a mock email service for testing
type MockService struct {
	SentEmails []SentEmail
}

// SentEmail represents a sent email for testing
type SentEmail struct {
	To      string
	Subject string
	Body    string
}

// Send mocks sending an email
func (m *MockService) Send(to, toName, subject, htmlContent, textContent string) error {
	m.SentEmails = append(m.SentEmails, SentEmail{
		To:      to,
		Subject: subject,
		Body:    htmlContent,
	})
	return nil
}
