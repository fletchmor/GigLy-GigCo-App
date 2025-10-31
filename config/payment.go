package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// PaymentConfig holds payment provider configurations
type PaymentConfig struct {
	Provider string
	Clover   CloverConfig
}

// CloverConfig holds Clover-specific configuration
type CloverConfig struct {
	Environment          string // sandbox or production
	MerchantID           string
	AccessToken          string
	APIAccessKey         string // PAKMS key for tokenization
	TokenizationEndpoint string
	APIEndpoint          string
	PAKMSEndpoint        string
	WebhookSecret        string
	PlatformFeePercent   float64 // Platform fee percentage (e.g., 10.0 for 10%)
}

var Payment *PaymentConfig

// InitPaymentConfig initializes payment configuration from environment variables
func InitPaymentConfig() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found for payment config, using environment variables")
	}

	environment := getEnvOrDefault("CLOVER_ENVIRONMENT", "sandbox")

	Payment = &PaymentConfig{
		Provider: getEnvOrDefault("PAYMENT_PROVIDER", "clover"),
		Clover: CloverConfig{
			Environment:          environment,
			MerchantID:           os.Getenv("CLOVER_MERCHANT_ID"),
			AccessToken:          os.Getenv("CLOVER_ACCESS_TOKEN"),
			APIAccessKey:         os.Getenv("CLOVER_API_ACCESS_KEY"),
			WebhookSecret:        os.Getenv("CLOVER_WEBHOOK_SECRET"),
			PlatformFeePercent:   parseFloatEnv("PLATFORM_FEE_PERCENT", 10.0),
			TokenizationEndpoint: getCloverEndpoint(environment, "tokenization"),
			APIEndpoint:          getCloverEndpoint(environment, "api"),
			PAKMSEndpoint:        getCloverEndpoint(environment, "pakms"),
		},
	}

	// Validate required Clover configuration
	if Payment.Provider == "clover" {
		if Payment.Clover.MerchantID == "" {
			log.Println("WARNING: CLOVER_MERCHANT_ID is not set")
		}
		if Payment.Clover.AccessToken == "" {
			log.Println("WARNING: CLOVER_ACCESS_TOKEN is not set")
		}
		if Payment.Clover.APIAccessKey == "" {
			log.Println("WARNING: CLOVER_API_ACCESS_KEY is not set - required for card tokenization")
		}
	}

	log.Printf("Payment config initialized: Provider=%s, Environment=%s",
		Payment.Provider, Payment.Clover.Environment)
}

// getCloverEndpoint returns the appropriate Clover endpoint based on environment
func getCloverEndpoint(environment, endpointType string) string {
	isSandbox := strings.ToLower(environment) == "sandbox"

	switch endpointType {
	case "tokenization":
		if isSandbox {
			return "https://token-sandbox.dev.clover.com/v1/tokens"
		}
		return "https://token.clover.com/v1/tokens"
	case "api":
		if isSandbox {
			return "https://scl-sandbox.dev.clover.com"
		}
		return "https://scl.clover.com"
	case "pakms":
		if isSandbox {
			return "https://scl-sandbox.dev.clover.com/pakms/apikey"
		}
		return "https://scl.clover.com/pakms/apikey"
	default:
		return ""
	}
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseFloatEnv parses a float environment variable or returns default
func parseFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var result float64
		if _, err := fmt.Sscanf(value, "%f", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

// CalculatePlatformFee calculates the platform fee based on amount
func (c *CloverConfig) CalculatePlatformFee(amount float64) float64 {
	return amount * (c.PlatformFeePercent / 100.0)
}

// CalculateProcessingFee calculates Clover's processing fee (typically 2.6% + $0.10)
func (c *CloverConfig) CalculateProcessingFee(amount float64) float64 {
	percentage := 2.6 // Clover's typical percentage
	fixedFee := 0.10   // Fixed fee per transaction
	return (amount * (percentage / 100.0)) + fixedFee
}

// CalculateNetAmount calculates the net amount after fees
func (c *CloverConfig) CalculateNetAmount(amount float64) (netAmount, platformFee, processingFee float64) {
	platformFee = c.CalculatePlatformFee(amount)
	processingFee = c.CalculateProcessingFee(amount)
	netAmount = amount - platformFee - processingFee
	return
}
