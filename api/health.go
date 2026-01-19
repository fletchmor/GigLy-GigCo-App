package api

import (
	"app/config"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"
)

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status      string                   `json:"status"`
	Version     string                   `json:"version,omitempty"`
	Environment string                   `json:"environment,omitempty"`
	Timestamp   time.Time                `json:"timestamp"`
	Uptime      string                   `json:"uptime,omitempty"`
	Checks      map[string]ComponentCheck `json:"checks,omitempty"`
}

// ComponentCheck represents a health check for a component
type ComponentCheck struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Latency  string `json:"latency,omitempty"`
}

var startTime = time.Now()

// Note: Basic HealthCheck is defined in api.go for backwards compatibility
// Use ReadinessCheck and LivenessCheck for Kubernetes-style health checks

// ReadinessCheck returns detailed health status with dependency checks
// @Summary Readiness check endpoint
// @Description Returns detailed health status including database and other dependencies
// @Tags health
// @Produce json
// @Success 200 {object} HealthStatus
// @Failure 503 {object} HealthStatus
// @Router /ready [get]
func ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]ComponentCheck)
	overallHealthy := true

	// Check database
	dbCheck := checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallHealthy = false
	}

	// Check Temporal (if configured)
	temporalHost := os.Getenv("TEMPORAL_HOST")
	if temporalHost != "" {
		temporalCheck := checkTemporal()
		checks["temporal"] = temporalCheck
		if temporalCheck.Status != "healthy" {
			overallHealthy = false
		}
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !overallHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	healthStatus := HealthStatus{
		Status:      status,
		Version:     os.Getenv("APP_VERSION"),
		Environment: os.Getenv("APP_ENV"),
		Timestamp:   time.Now(),
		Uptime:      time.Since(startTime).String(),
		Checks:      checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(healthStatus)
}

// LivenessCheck returns if the application is running
// @Summary Liveness check endpoint
// @Description Returns if the application process is alive
// @Tags health
// @Produce json
// @Success 200 {object} HealthStatus
// @Router /live [get]
func LivenessCheck(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "alive",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// MetricsCheck returns application metrics
// @Summary Metrics endpoint
// @Description Returns application runtime metrics
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics [get]
func MetricsCheck(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime).String(),
		"runtime": map[string]interface{}{
			"goroutines":    runtime.NumGoroutine(),
			"cpus":          runtime.NumCPU(),
			"go_version":    runtime.Version(),
		},
		"memory": map[string]interface{}{
			"alloc_mb":       memStats.Alloc / 1024 / 1024,
			"total_alloc_mb": memStats.TotalAlloc / 1024 / 1024,
			"sys_mb":         memStats.Sys / 1024 / 1024,
			"num_gc":         memStats.NumGC,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// checkDatabase verifies database connectivity
func checkDatabase() ComponentCheck {
	if config.DB == nil {
		return ComponentCheck{
			Status:  "unhealthy",
			Message: "database connection not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := config.DB.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		return ComponentCheck{
			Status:  "unhealthy",
			Message: "database ping failed: " + err.Error(),
			Latency: latency.String(),
		}
	}

	return ComponentCheck{
		Status:  "healthy",
		Message: "database connection OK",
		Latency: latency.String(),
	}
}

// checkTemporal verifies Temporal connectivity
func checkTemporal() ComponentCheck {
	// For now, just check if the host is configured
	// In a full implementation, you would ping the Temporal service
	temporalHost := os.Getenv("TEMPORAL_HOST")
	if temporalHost == "" {
		return ComponentCheck{
			Status:  "unknown",
			Message: "temporal host not configured",
		}
	}

	// TODO: Add actual Temporal health check
	// This would require importing the Temporal client and pinging the service
	return ComponentCheck{
		Status:  "healthy",
		Message: "temporal configured at " + temporalHost,
	}
}
