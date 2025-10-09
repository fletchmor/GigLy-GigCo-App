package api

import (
	"app/config"
	"app/internal/auth"
	"app/internal/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Password     string   `json:"password"`
	Phone        string   `json:"phone,omitempty"`
	Address      string   `json:"address"`
	Role         string   `json:"role"`
	Latitude     float64  `json:"latitude,omitempty"`
	Longitude    float64  `json:"longitude,omitempty"`
	PlaceID      string   `json:"place_id,omitempty"`
	Skills       []string `json:"skills,omitempty"`       // For gig workers
	Availability string   `json:"availability,omitempty"` // For gig workers
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	ID            int       `json:"id"`
	UUID          string    `json:"uuid"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone,omitempty"`
	Address       string    `json:"address"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	CreatedAt     time.Time `json:"created_at"`
	Token         string    `json:"token,omitempty"` // JWT token (placeholder for now)
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	ID            int       `json:"id"`
	UUID          string    `json:"uuid"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	CreatedAt     time.Time `json:"created_at"`
	Token         string    `json:"token"`
}

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Phone validation regex (basic US format)
var phoneRegex = regexp.MustCompile(`^\+?1?[-.\s]?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}$`)

// RegisterUser handles user registration with role selection
// RegisterUser godoc
// @Summary Register a new user
// @Description Register a new user in the system with email, name and role
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Registration data"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} map[string]string "Invalid request data"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateRegistrationRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Normalize and clean data
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))

	// Clean phone number if provided
	if req.Phone != "" {
		req.Phone = cleanPhoneNumber(req.Phone)
	}

	// Check if email already exists
	var existingID int
	checkQuery := "SELECT id FROM people WHERE email = $1"
	err = config.DB.QueryRow(checkQuery, req.Email).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			http.Error(w, "Email address already registered", http.StatusConflict)
			return
		}
		log.Printf("Database error checking existing email: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert new user into people table
	insertQuery := `
		INSERT INTO people (
			email, name, password_hash, phone, address, latitude, longitude, place_id,
			role, is_active, email_verified, phone_verified, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, uuid, created_at`

	var response RegisterResponse
	now := time.Now()

	// Set default values
	isActive := true
	emailVerified := false
	phoneVerified := false

	// For development environment, auto-verify emails ending with @gigco.dev
	if strings.HasSuffix(req.Email, "@gigco.dev") {
		emailVerified = true
	}

	err = config.DB.QueryRow(
		insertQuery,
		req.Email,
		req.Name,
		string(hashedPassword),
		nullString(req.Phone),
		req.Address,
		nullFloat64(req.Latitude),
		nullFloat64(req.Longitude),
		nullString(req.PlaceID),
		req.Role,
		isActive,
		emailVerified,
		phoneVerified,
		now,
		now,
	).Scan(&response.ID, &response.UUID, &response.CreatedAt)

	if err != nil {
		// Handle unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "email") {
					http.Error(w, "Email address already registered", http.StatusConflict)
					return
				}
			case "23514": // check_violation
				http.Error(w, "Invalid role specified", http.StatusBadRequest)
				return
			}
		}
		log.Printf("Database error inserting user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create default notification preferences for the new user
	err = createDefaultNotificationPreferences(response.ID)
	if err != nil {
		log.Printf("Warning: Failed to create notification preferences for user %d: %v", response.ID, err)
		// Don't fail the registration for this
	}

	// Build response
	response.Name = req.Name
	response.Email = req.Email
	response.Phone = req.Phone
	response.Address = req.Address
	response.Role = req.Role
	response.IsActive = isActive
	response.EmailVerified = emailVerified
	response.PhoneVerified = phoneVerified

	// Generate JWT token
	token, err := auth.GenerateJWT(response.ID, response.UUID, response.Email, response.Role)
	if err != nil {
		log.Printf("Failed to generate JWT token: %v", err)
		// Don't fail registration for token generation error
	} else {
		response.Token = token
	}

	// Log successful registration
	log.Printf("New user registered: ID=%d, Email=%s, Role=%s", response.ID, response.Email, response.Role)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// validateRegistrationRequest validates the registration request
func validateRegistrationRequest(req *RegisterRequest) error {
	// Required fields
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if req.Address == "" {
		return fmt.Errorf("address is required")
	}
	if req.Role == "" {
		return fmt.Errorf("role is required")
	}

	// Validate password strength
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	// Validate email format
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate role
	validRoles := []string{"consumer", "gig_worker", "admin"}
	roleValid := false
	for _, validRole := range validRoles {
		if req.Role == validRole {
			roleValid = true
			break
		}
	}
	if !roleValid {
		return fmt.Errorf("role must be one of: consumer, gig_worker, admin")
	}

	// Validate phone if provided
	if req.Phone != "" && !phoneRegex.MatchString(req.Phone) {
		return fmt.Errorf("invalid phone number format")
	}

	// Validate name length
	if len(req.Name) < 2 || len(req.Name) > 255 {
		return fmt.Errorf("name must be between 2 and 255 characters")
	}

	// Validate email length
	if len(req.Email) > 255 {
		return fmt.Errorf("email must be less than 255 characters")
	}

	// Role-specific validations
	if req.Role == "admin" {
		// Only allow admin creation in development
		// In production, this should be restricted
		log.Printf("Warning: Admin user registration attempted for email: %s", req.Email)
	}

	return nil
}

// createDefaultNotificationPreferences creates default notification preferences for a new user
func createDefaultNotificationPreferences(userID int) error {
	// Get all notification types
	notificationTypes := []string{
		"job_posted", "job_accepted", "job_completed",
		"payment_received", "system_message",
	}

	for _, notificationType := range notificationTypes {
		insertQuery := `
			INSERT INTO notification_preferences (
				user_id, type, email_enabled, push_enabled, sms_enabled
			) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id, type) DO NOTHING`

		// Default preferences
		emailEnabled := true
		pushEnabled := true
		smsEnabled := false

		// Customize defaults by notification type
		if notificationType == "system_message" {
			pushEnabled = false // System messages via email only by default
		}

		_, err := config.DB.Exec(
			insertQuery,
			userID,
			notificationType,
			emailEnabled,
			pushEnabled,
			smsEnabled,
		)
		if err != nil {
			return fmt.Errorf("failed to create notification preference for %s: %v", notificationType, err)
		}
	}

	return nil
}

// Helper functions for handling nullable database fields
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func customNullString(s string) model.NullString {
	if s == "" {
		return model.NullString{NullString: sql.NullString{Valid: false}}
	}
	return model.NullString{NullString: sql.NullString{String: s, Valid: true}}
}

func nullFloat64(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

// cleanPhoneNumber removes formatting from phone numbers
func cleanPhoneNumber(phone string) string {
	// Remove common formatting characters
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// Add +1 if it's a 10-digit US number
	if len(cleaned) == 10 && !strings.HasPrefix(cleaned, "+") {
		cleaned = "+1" + cleaned
	}

	return cleaned
}

// LoginUser godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string "Invalid request data"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest

	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if loginReq.Email == "" || loginReq.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Get user with password hash
	var user model.User
	var passwordHash sql.NullString
	query := `
		SELECT id, uuid, name, email, role, is_active, email_verified, phone_verified, created_at, password_hash
		FROM people WHERE email = $1 AND is_active = true
	`

	err = config.DB.QueryRow(query, strings.ToLower(strings.TrimSpace(loginReq.Email))).Scan(
		&user.ID, &user.Uuid, &user.Name, &user.Email, &user.Role,
		&user.IsActive, &user.EmailVerified, &user.PhoneVerified, &user.CreatedAt, &passwordHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		log.Printf("Database error during login: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if !passwordHash.Valid || passwordHash.String == "" {
		// No password set, reject login
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash.String), []byte(loginReq.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID, user.Uuid, user.Email, user.Role)
	if err != nil {
		log.Printf("Failed to generate JWT token: %v", err)
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		ID:            user.ID,
		UUID:          user.Uuid,
		Name:          user.Name,
		Email:         user.Email,
		Role:          user.Role,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
		CreatedAt:     user.CreatedAt,
		Token:         token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// LogoutUser handles user logout
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// NOTE: JWT token blacklisting would require a cache/database implementation
	// For stateless JWT, logout is handled client-side by removing the token

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	})
}

// RefreshToken refreshes the JWT token
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var refreshReq struct {
		Token string `json:"token"`
	}

	err := json.NewDecoder(r.Body).Decode(&refreshReq)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if refreshReq.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Validate and refresh JWT token
	newToken, err := auth.RefreshJWT(refreshReq.Token)
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   newToken,
	})
}

// VerifyEmail verifies a user's email address
func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var verifyReq struct {
		Token string `json:"token"`
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&verifyReq)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if verifyReq.Token == "" || verifyReq.Email == "" {
		http.Error(w, "Token and email are required", http.StatusBadRequest)
		return
	}

	// TODO: Validate verification token
	// For now, just update the email_verified status

	query := "UPDATE people SET email_verified = true, updated_at = NOW() WHERE email = $1"
	_, err = config.DB.Exec(query, strings.ToLower(strings.TrimSpace(verifyReq.Email)))
	if err != nil {
		log.Printf("Database error verifying email: %v", err)
		http.Error(w, "Failed to verify email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Email verified successfully",
	})
}

// ForgotPassword initiates password reset process
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var forgotReq struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&forgotReq)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if forgotReq.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if user exists
	var userID int
	query := "SELECT id FROM people WHERE email = $1 AND is_active = true"
	err = config.DB.QueryRow(query, strings.ToLower(strings.TrimSpace(forgotReq.Email))).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists, return success anyway
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "If the email exists, a password reset link has been sent",
			})
			return
		}
		log.Printf("Database error during password reset: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: Generate password reset token and send email
	// For now, just log and return success
	log.Printf("Password reset requested for user ID: %d", userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword resets user password with token
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resetReq struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	err := json.NewDecoder(r.Body).Decode(&resetReq)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if resetReq.Token == "" || resetReq.NewPassword == "" {
		http.Error(w, "Token and new password are required", http.StatusBadRequest)
		return
	}

	// Validate password strength
	if len(resetReq.NewPassword) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// TODO: Validate reset token and update password hash
	// For now, just return success
	log.Printf("Password reset completed for token: %s", resetReq.Token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password reset successfully",
	})
}
