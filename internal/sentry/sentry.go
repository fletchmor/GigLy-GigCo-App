package sentry

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// Config holds Sentry configuration
type Config struct {
	DSN              string
	Environment      string
	Release          string
	Debug            bool
	SampleRate       float64
	TracesSampleRate float64
}

// Init initializes Sentry error tracking
func Init(cfg Config) error {
	if cfg.DSN == "" {
		return fmt.Errorf("sentry DSN is required")
	}

	sampleRate := cfg.SampleRate
	if sampleRate == 0 {
		sampleRate = 1.0
	}

	tracesSampleRate := cfg.TracesSampleRate
	if tracesSampleRate == 0 {
		tracesSampleRate = 0.2 // Sample 20% of transactions
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		Debug:            cfg.Debug,
		SampleRate:       sampleRate,
		TracesSampleRate: tracesSampleRate,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Scrub sensitive data
			scrubSensitiveData(event)
			return event
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize sentry: %w", err)
	}

	return nil
}

// InitFromEnv initializes Sentry from environment variables
func InitFromEnv() error {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		// Sentry is optional - just log and return
		return nil
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	release := os.Getenv("APP_VERSION")
	if release == "" {
		release = "unknown"
	}

	return Init(Config{
		DSN:              dsn,
		Environment:      env,
		Release:          release,
		Debug:            env != "production",
		SampleRate:       1.0,
		TracesSampleRate: 0.2,
	})
}

// Middleware returns HTTP middleware for Sentry
func Middleware() func(http.Handler) http.Handler {
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	})

	return func(next http.Handler) http.Handler {
		return sentryHandler.Handle(next)
	}
}

// CaptureError captures an error to Sentry
func CaptureError(err error) {
	sentry.CaptureException(err)
}

// CaptureErrorWithContext captures an error with additional context
func CaptureErrorWithContext(err error, context map[string]interface{}) {
	sentry.WithScope(func(scope *sentry.Scope) {
		for key, value := range context {
			scope.SetExtra(key, value)
		}
		sentry.CaptureException(err)
	})
}

// CaptureMessage captures a message to Sentry
func CaptureMessage(message string) {
	sentry.CaptureMessage(message)
}

// SetUser sets user information for Sentry events
func SetUser(userID string, email string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:    userID,
			Email: email,
		})
	})
}

// SetTag sets a tag for Sentry events
func SetTag(key, value string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

// Flush flushes any pending events
func Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

// RecoverWithSentry recovers from panics and reports to Sentry
func RecoverWithSentry() {
	if err := recover(); err != nil {
		sentry.CurrentHub().Recover(err)
		sentry.Flush(2 * time.Second)
		panic(err) // Re-panic after reporting
	}
}

// scrubSensitiveData removes sensitive information from Sentry events
func scrubSensitiveData(event *sentry.Event) {
	sensitiveKeys := []string{
		"password", "token", "secret", "api_key", "apikey",
		"authorization", "auth", "credential", "credit_card",
		"card_number", "cvv", "ssn", "social_security",
	}

	// Scrub request headers
	if event.Request != nil {
		for _, key := range sensitiveKeys {
			delete(event.Request.Headers, key)
		}
		// Clear request body data as it may contain sensitive info
		event.Request.Data = ""
	}

	// Scrub extra data
	for _, key := range sensitiveKeys {
		delete(event.Extra, key)
	}
}
