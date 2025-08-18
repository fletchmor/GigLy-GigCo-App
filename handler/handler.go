package handler

import (
	"app/api"
	"app/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func GetHandlers(r chi.Router) {
	r.Get("/health", api.HealthCheck)
	r.Get("/", middleware.ServeEmailForm)
	r.Get("/email-submit", middleware.HandleEmailSubmission)
	
	// User Management
	r.Get("/api/v1/customers/{id}", api.GetCustomerByID)
	r.Get("/api/v1/users/profile", api.GetUserProfile)
	r.Get("/api/v1/users/{id}", api.GetUserByID)
	
	// GigWorker Management
	r.Get("/api/v1/gigworkers", api.GetGigWorkers)
	r.Get("/api/v1/gigworkers/{id}", api.GetGigWorkerByID)
	
	// Job Management
	r.Get("/api/v1/jobs", api.GetJobs)
	r.Get("/api/v1/jobs/{id}", api.GetJobByID)
	r.Get("/api/v1/jobs/my-jobs", api.GetMyJobs)
	r.Get("/api/v1/jobs/available", api.GetAvailableJobs)
}

func PostHandlers(r chi.Router) {
	// Authentication
	r.Post("/api/v1/auth/register", api.RegisterUser)
	r.Post("/api/v1/auth/login", api.LoginUser)
	r.Post("/api/v1/auth/logout", api.LogoutUser)
	r.Post("/api/v1/auth/refresh", api.RefreshToken)
	r.Post("/api/v1/auth/verify-email", api.VerifyEmail)
	r.Post("/api/v1/auth/forgot-password", api.ForgotPassword)
	r.Post("/api/v1/auth/reset-password", api.ResetPassword)
	
	// User Management
	r.Post("/api/v1/users/create", api.CreateUser)
	
	// GigWorker Management
	r.Post("/api/v1/gigworkers/create", api.CreateGigWorker)
	
	// Job Management
	r.Post("/api/v1/jobs/create", api.CreateJob)
	r.Post("/api/v1/jobs/{id}/accept", api.AcceptJob)
	r.Post("/api/v1/jobs/{id}/send-offer", api.SendJobOffer)
	
	// Job Workflow endpoints
	r.Post("/api/v1/jobs/{id}/accept-offer", api.AcceptJobOffer)
	r.Post("/api/v1/jobs/{id}/reject-offer", api.RejectJobOffer)
	r.Post("/api/v1/jobs/{id}/start", api.StartJob)
	r.Post("/api/v1/jobs/{id}/complete", api.CompleteJob)
	r.Post("/api/v1/jobs/{id}/review", api.SubmitReview)
	
	// Schedule Management
	r.Post("/api/v1/schedules/create", api.CreateSchedule)
	
	// Transaction Management
	r.Post("/api/v1/transactions/create", api.CreateTransaction)
}

func PutHandlers(r chi.Router) {
	// User Management
	r.Put("/api/v1/users/profile", api.UpdateUserProfile)
	r.Put("/api/v1/users/{id}", api.UpdateUser)
	
	// GigWorker Management
	r.Put("/api/v1/gigworkers/{id}", api.UpdateGigWorker)
	
	// Job Management
	r.Put("/api/v1/jobs/{id}", api.UpdateJob)
}

func DeleteHandlers(r chi.Router) {
	// User Management
	r.Delete("/api/v1/users/{id}", api.DeactivateUser)
	
	// GigWorker Management
	r.Delete("/api/v1/gigworkers/{id}", api.DeactivateGigWorker)
	
	// Job Management
	r.Delete("/api/v1/jobs/{id}", api.CancelJob)
}
