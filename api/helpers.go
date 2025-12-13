// Package api provides HTTP API handlers and utilities for the GigCo platform.
// This file contains helper functions for common API operations including
// parameter parsing, validation, and standardized JSON responses.
package api

import (
	"app/internal/model"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// API Constants
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)

// Query parameter validation helpers

// ParseIntParam parses an integer query parameter with validation
func ParseIntParam(r *http.Request, paramName string, defaultValue int, min int, max int) (int, error) {
	valueStr := r.URL.Query().Get(paramName)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, &ValidationError{
			Field:   paramName,
			Message: "must be a valid integer",
			Value:   valueStr,
		}
	}

	if value < min {
		return 0, &ValidationError{
			Field:   paramName,
			Message: "must be at least " + strconv.Itoa(min),
			Value:   valueStr,
		}
	}

	if max > 0 && value > max {
		return 0, &ValidationError{
			Field:   paramName,
			Message: "must not exceed " + strconv.Itoa(max),
			Value:   valueStr,
		}
	}

	return value, nil
}

// ParseBoolParam parses a boolean query parameter
func ParseBoolParam(r *http.Request, paramName string) (bool, bool, error) {
	valueStr := r.URL.Query().Get(paramName)
	if valueStr == "" {
		return false, false, nil // not provided
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false, true, &ValidationError{
			Field:   paramName,
			Message: "must be 'true' or 'false'",
			Value:   valueStr,
		}
	}

	return value, true, nil
}

// ParseDateParam parses a date query parameter (RFC3339 format)
func ParseDateParam(r *http.Request, paramName string) (*time.Time, error) {
	valueStr := r.URL.Query().Get(paramName)
	if valueStr == "" {
		return nil, nil
	}

	// Try multiple date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"2006-01-02 15:04:05",
	}

	var parsedTime time.Time
	var err error
	for _, format := range formats {
		parsedTime, err = time.Parse(format, valueStr)
		if err == nil {
			return &parsedTime, nil
		}
	}

	return nil, &ValidationError{
		Field:   paramName,
		Message: "invalid date format (use YYYY-MM-DD or RFC3339)",
		Value:   valueStr,
	}
}

// Error response helpers

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// RespondWithError sends a JSON error response
func RespondWithError(w http.ResponseWriter, statusCode int, errorMsg string) {
	RespondWithJSON(w, statusCode, model.ErrorResponse{
		Error: errorMsg,
	})
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(w http.ResponseWriter, err *ValidationError) {
	RespondWithJSON(w, http.StatusBadRequest, model.ErrorResponse{
		Error:   "Validation failed",
		Message: err.Error(),
		Code:    "VALIDATION_ERROR",
		Details: map[string]string{
			err.Field: err.Message,
		},
	})
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, log it but don't try to send another response
		// as headers are already written
		return
	}
}
