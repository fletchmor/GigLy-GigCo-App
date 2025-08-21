package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret       []byte
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims represents the claims structure for JWT tokens
type JWTClaims struct {
	UserID int    `json:"user_id"`
	UUID   string `json:"uuid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// InitJWT initializes the JWT secret key
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Generate a random secret in development
		log.Println("Warning: JWT_SECRET not set, generating random secret for development")
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			log.Fatal("Failed to generate JWT secret:", err)
		}
		secret = hex.EncodeToString(randomBytes)
		log.Printf("Generated JWT secret: %s", secret)
	}
	jwtSecret = []byte(secret)
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(userID int, uuid, email, role string) (string, error) {
	if len(jwtSecret) == 0 {
		InitJWT()
	}

	expirationTime := time.Now().Add(24 * time.Hour) // 24 hours

	claims := &JWTClaims{
		UserID: userID,
		UUID:   uuid,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gigco-api",
			Subject:   strconv.Itoa(userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	if len(jwtSecret) == 0 {
		InitJWT()
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshJWT creates a new token from an existing valid token
func RefreshJWT(tokenString string) (string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is close to expiry (refresh if less than 1 hour remaining)
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return tokenString, nil // Token is still fresh, return original
	}

	// Generate new token with same claims but updated expiry
	return GenerateJWT(claims.UserID, claims.UUID, claims.Email, claims.Role)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateResetToken generates a secure random token for password reset
func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateVerificationToken generates a secure random token for email verification
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
