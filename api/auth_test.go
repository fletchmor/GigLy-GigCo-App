package api

import (
	"testing"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid strong password",
			password: "MyStr0ng!Pass",
			wantErr:  false,
		},
		{
			name:     "valid with all types",
			password: "Abcd1234!@#$",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Short1!",
			wantErr:  true,
			errMsg:   "at least 10 characters",
		},
		{
			name:     "no uppercase but has 3 types",
			password: "lowercase123!",
			wantErr:  false, // Has 3 types: lower, number, special - should pass
		},
		{
			name:     "no lowercase",
			password: "UPPERCASE123!",
			wantErr:  false, // Has 3 types: upper, number, special
		},
		{
			name:     "no numbers",
			password: "NoNumbersHere!",
			wantErr:  false, // Has 3 types: upper, lower, special
		},
		{
			name:     "no special chars",
			password: "NoSpecial123",
			wantErr:  false, // Has 3 types: upper, lower, number
		},
		{
			name:     "only lowercase and numbers",
			password: "onlylowercase123",
			wantErr:  true,
			errMsg:   "3 of",
		},
		{
			name:     "common password",
			password: "Password123!",
			wantErr:  false, // Not in our common list
		},
		{
			name:     "common weak password - password123",
			password: "password123",
			wantErr:  true, // In common list and lacks complexity
		},
		{
			name:     "exactly 10 characters valid",
			password: "Abcd1234!@",
			wantErr:  false,
		},
		{
			name:     "9 characters too short",
			password: "Abcd123!@",
			wantErr:  true,
			errMsg:   "at least 10 characters",
		},
		{
			name:     "max length 72 characters",
			password: "Abcd1234!" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr:  true,
			errMsg:   "less than 72 characters",
		},
		{
			name:     "unicode characters",
			password: "Pässwörd123!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("validatePasswordStrength() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateRegistrationRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid consumer registration",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "consumer",
			},
			wantErr: false,
		},
		{
			name: "valid worker registration",
			req: RegisterRequest{
				Name:     "Jane Worker",
				Email:    "jane@example.com",
				Password: "W0rkerP@ss123!",
				Address:  "456 Work Ave",
				Role:     "gig_worker",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: RegisterRequest{
				Name:     "",
				Email:    "test@example.com",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "consumer",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing email",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "consumer",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email format",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "consumer",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "missing password",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "",
				Address:  "123 Main St",
				Role:     "consumer",
			},
			wantErr: true,
			errMsg:  "password is required",
		},
		{
			name: "weak password",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "weak",
				Address:  "123 Main St",
				Role:     "consumer",
			},
			wantErr: true,
			errMsg:  "at least 10 characters",
		},
		{
			name: "missing address",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Str0ngP@ssword!",
				Address:  "",
				Role:     "consumer",
			},
			wantErr: true,
			errMsg:  "address is required",
		},
		{
			name: "missing role",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "",
			},
			wantErr: true,
			errMsg:  "role is required",
		},
		{
			name: "invalid role",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "invalid_role",
			},
			wantErr: true,
			errMsg:  "role must be one of",
		},
		{
			name: "name too short",
			req: RegisterRequest{
				Name:     "J",
				Email:    "john@example.com",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "consumer",
			},
			wantErr: true,
			errMsg:  "between 2 and 255 characters",
		},
		{
			name: "valid phone number",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "consumer",
				Phone:    "555-123-4567",
			},
			wantErr: false,
		},
		{
			name: "invalid phone number",
			req: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "Str0ngP@ssword!",
				Address:  "123 Main St",
				Role:     "consumer",
				Phone:    "invalid",
			},
			wantErr: true,
			errMsg:  "invalid phone number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegistrationRequest(&tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRegistrationRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("validateRegistrationRequest() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestCleanPhoneNumber(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{
			name:  "already clean with country code",
			phone: "+15551234567",
			want:  "+15551234567",
		},
		{
			name:  "10 digit adds country code",
			phone: "5551234567",
			want:  "+15551234567",
		},
		{
			name:  "dashes removed",
			phone: "555-123-4567",
			want:  "+15551234567",
		},
		{
			name:  "parentheses and spaces removed",
			phone: "(555) 123-4567",
			want:  "+15551234567",
		},
		{
			name:  "dots removed",
			phone: "555.123.4567",
			want:  "+15551234567",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanPhoneNumber(tt.phone)
			if got != tt.want {
				t.Errorf("cleanPhoneNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
