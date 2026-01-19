package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	// Log is the global logger instance
	Log zerolog.Logger
)

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Pretty     bool   // Use console writer for pretty output
	TimeFormat string // Time format string
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Pretty:     false,
		TimeFormat: time.RFC3339,
	}
}

// Init initializes the global logger
func Init(cfg Config) {
	// Set time format
	zerolog.TimeFieldFormat = cfg.TimeFormat

	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Choose output writer
	var writer io.Writer = os.Stdout
	if cfg.Pretty {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger with common fields
	Log = zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Caller().
		Str("service", "gigco-api").
		Logger()
}

// InitFromEnv initializes logger from environment variables
func InitFromEnv() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	env := os.Getenv("APP_ENV")
	pretty := env != "production"

	Init(Config{
		Level:      level,
		Pretty:     pretty,
		TimeFormat: time.RFC3339,
	})
}

// WithRequestID returns a logger with request ID
func WithRequestID(requestID string) zerolog.Logger {
	return Log.With().Str("request_id", requestID).Logger()
}

// WithUserID returns a logger with user ID
func WithUserID(userID int) zerolog.Logger {
	return Log.With().Int("user_id", userID).Logger()
}

// WithJob returns a logger with job context
func WithJob(jobID int, jobUUID string) zerolog.Logger {
	return Log.With().
		Int("job_id", jobID).
		Str("job_uuid", jobUUID).
		Logger()
}

// Debug logs a debug message
func Debug() *zerolog.Event {
	return Log.Debug()
}

// Info logs an info message
func Info() *zerolog.Event {
	return Log.Info()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return Log.Warn()
}

// Error logs an error message
func Error() *zerolog.Event {
	return Log.Error()
}

// Fatal logs a fatal message and exits
func Fatal() *zerolog.Event {
	return Log.Fatal()
}
