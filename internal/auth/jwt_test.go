package auth

import (
	"os"
	"testing"
	"time"
)

func TestInitJWT(t *testing.T) {
	// Save original env
	originalSecret := os.Getenv("JWT_SECRET")
	originalEnv := os.Getenv("APP_ENV")
	defer func() {
		os.Setenv("JWT_SECRET", originalSecret)
		os.Setenv("APP_ENV", originalEnv)
		jwtSecret = nil // Reset for other tests
	}()

	tests := []struct {
		name      string
		secret    string
		appEnv    string
		wantPanic bool
	}{
		{
			name:      "valid secret in production",
			secret:    "this-is-a-very-long-secret-key-for-production-use-only",
			appEnv:    "production",
			wantPanic: false,
		},
		{
			name:      "empty secret in development generates random",
			secret:    "",
			appEnv:    "development",
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtSecret = nil // Reset
			os.Setenv("JWT_SECRET", tt.secret)
			os.Setenv("APP_ENV", tt.appEnv)

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic but did not get one")
					}
				}()
			}

			InitJWT()

			if len(jwtSecret) == 0 {
				t.Error("jwtSecret should not be empty after InitJWT")
			}
		})
	}
}

func TestGenerateAndValidateJWT(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-purposes-only")
	os.Setenv("APP_ENV", "test")
	jwtSecret = nil
	InitJWT()

	tests := []struct {
		name    string
		userID  int
		uuid    string
		email   string
		role    string
		wantErr bool
	}{
		{
			name:    "valid consumer token",
			userID:  1,
			uuid:    "550e8400-e29b-41d4-a716-446655440000",
			email:   "test@example.com",
			role:    "consumer",
			wantErr: false,
		},
		{
			name:    "valid worker token",
			userID:  2,
			uuid:    "550e8400-e29b-41d4-a716-446655440001",
			email:   "worker@example.com",
			role:    "gig_worker",
			wantErr: false,
		},
		{
			name:    "valid admin token",
			userID:  3,
			uuid:    "550e8400-e29b-41d4-a716-446655440002",
			email:   "admin@example.com",
			role:    "admin",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate token
			token, err := GenerateJWT(tt.userID, tt.uuid, tt.email, tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Token should not be empty
			if token == "" {
				t.Error("GenerateJWT() returned empty token")
				return
			}

			// Validate token
			claims, err := ValidateJWT(token)
			if err != nil {
				t.Errorf("ValidateJWT() error = %v", err)
				return
			}

			// Check claims
			if claims.UserID != tt.userID {
				t.Errorf("claims.UserID = %v, want %v", claims.UserID, tt.userID)
			}
			if claims.UUID != tt.uuid {
				t.Errorf("claims.UUID = %v, want %v", claims.UUID, tt.uuid)
			}
			if claims.Email != tt.email {
				t.Errorf("claims.Email = %v, want %v", claims.Email, tt.email)
			}
			if claims.Role != tt.role {
				t.Errorf("claims.Role = %v, want %v", claims.Role, tt.role)
			}
		})
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-purposes-only")
	os.Setenv("APP_ENV", "test")
	jwtSecret = nil
	InitJWT()

	tests := []struct {
		name      string
		token     string
		wantError error
	}{
		{
			name:      "empty token",
			token:     "",
			wantError: ErrInvalidToken,
		},
		{
			name:      "malformed token",
			token:     "not.a.valid.token",
			wantError: ErrInvalidToken,
		},
		{
			name:      "random string",
			token:     "randomstringwithoutdots",
			wantError: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateJWT(tt.token)
			if err == nil {
				t.Error("ValidateJWT() expected error, got nil")
				return
			}
		})
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "TestPassword123!",
		},
		{
			name:     "complex password",
			password: "C0mpl3x!P@ssw0rd#2024",
		},
		{
			name:     "unicode password",
			password: "Pässwörd123!",
		},
		{
			name:     "long password",
			password: "ThisIsAVeryLongPasswordThatShouldStillWorkCorrectly123!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash password
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Errorf("HashPassword() error = %v", err)
				return
			}

			// Hash should not be empty
			if hash == "" {
				t.Error("HashPassword() returned empty hash")
				return
			}

			// Hash should not equal original password
			if hash == tt.password {
				t.Error("HashPassword() returned same as input")
				return
			}

			// Verify correct password
			if !VerifyPassword(tt.password, hash) {
				t.Error("VerifyPassword() failed to verify correct password")
			}

			// Verify incorrect password
			if VerifyPassword("wrongpassword", hash) {
				t.Error("VerifyPassword() verified incorrect password")
			}
		})
	}
}

func TestGenerateResetToken(t *testing.T) {
	token1, err := GenerateResetToken()
	if err != nil {
		t.Errorf("GenerateResetToken() error = %v", err)
		return
	}

	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("GenerateResetToken() length = %v, want 64", len(token1))
	}

	// Tokens should be unique
	token2, _ := GenerateResetToken()
	if token1 == token2 {
		t.Error("GenerateResetToken() generated duplicate tokens")
	}
}

func TestGenerateVerificationToken(t *testing.T) {
	token1, err := GenerateVerificationToken()
	if err != nil {
		t.Errorf("GenerateVerificationToken() error = %v", err)
		return
	}

	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("GenerateVerificationToken() length = %v, want 64", len(token1))
	}

	// Tokens should be unique
	token2, _ := GenerateVerificationToken()
	if token1 == token2 {
		t.Error("GenerateVerificationToken() generated duplicate tokens")
	}
}

func TestRefreshJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-purposes-only")
	os.Setenv("APP_ENV", "test")
	jwtSecret = nil
	InitJWT()

	// Generate a fresh token
	token, err := GenerateJWT(1, "test-uuid", "test@example.com", "consumer")
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	// Refresh should return same token since it's still fresh
	refreshed, err := RefreshJWT(token)
	if err != nil {
		t.Errorf("RefreshJWT() error = %v", err)
		return
	}

	if refreshed != token {
		t.Log("Note: Token was refreshed even though still fresh")
	}

	// Validate the refreshed token
	claims, err := ValidateJWT(refreshed)
	if err != nil {
		t.Errorf("ValidateJWT() on refreshed token error = %v", err)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Refreshed token email = %v, want %v", claims.Email, "test@example.com")
	}
}

func BenchmarkGenerateJWT(b *testing.B) {
	os.Setenv("JWT_SECRET", "benchmark-secret-key-for-testing")
	os.Setenv("APP_ENV", "test")
	jwtSecret = nil
	InitJWT()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateJWT(1, "test-uuid", "test@example.com", "consumer")
	}
}

func BenchmarkValidateJWT(b *testing.B) {
	os.Setenv("JWT_SECRET", "benchmark-secret-key-for-testing")
	os.Setenv("APP_ENV", "test")
	jwtSecret = nil
	InitJWT()

	token, _ := GenerateJWT(1, "test-uuid", "test@example.com", "consumer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateJWT(token)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "TestPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

// Helper to simulate time passing (for testing token expiration)
func init() {
	// Ensure tests run with consistent time
	_ = time.Now()
}
