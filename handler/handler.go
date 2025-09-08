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

	// User Management - Protected endpoints
	r.With(middleware.RequireRoles("admin", "consumer")).Get("/api/v1/customers/{id}", api.GetCustomerByID)
	r.Get("/api/v1/users/profile", api.GetUserProfile) // Any authenticated user
	r.With(middleware.RequireRole("admin")).Get("/api/v1/users/{id}", api.GetUserByID)

	// GigWorker Management
	r.With(middleware.RequireRoles("admin", "consumer")).Get("/api/v1/gigworkers", api.GetGigWorkers)
	r.Get("/api/v1/gigworkers/{id}", api.GetGigWorkerByID) // Any authenticated user

	// Job Management
	r.Get("/api/v1/jobs", api.GetJobs)           // Any authenticated user
	r.Get("/api/v1/jobs/{id}", api.GetJobByID)   // Any authenticated user
	r.Get("/api/v1/jobs/my-jobs", api.GetMyJobs) // Any authenticated user
	r.With(middleware.RequireRole("gig_worker")).Get("/api/v1/jobs/available", api.GetAvailableJobs)

	// Review Management
	r.Get("/api/v1/reviews", api.GetReviews)            // Any authenticated user (public reviews only)
	r.Get("/api/v1/reviews/{id}", api.GetReviewByID)    // Any authenticated user
	r.Get("/api/v1/jobs/{id}/reviews", api.GetJobReviews) // Any authenticated user
	r.Get("/api/v1/users/{id}/reviews", api.GetUserReviewStats) // Any authenticated user
	r.Get("/api/v1/reviews/stats", api.GetPlatformReviewStats) // Any authenticated user
	r.Get("/api/v1/reviews/top-rated", api.GetTopRatedUsers) // Any authenticated user
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

	// User Management - Protected endpoints
	r.With(middleware.RequireRole("admin")).Post("/api/v1/users/create", api.CreateUser)

	// GigWorker Management
	r.Post("/api/v1/gigworkers/create", api.CreateGigWorker) // Any authenticated user can register as gig worker

	// Job Management
	r.With(middleware.RequireRoles("admin", "consumer")).Post("/api/v1/jobs/create", api.CreateJob)
	r.With(middleware.RequireRole("gig_worker")).Post("/api/v1/jobs/{id}/accept", api.AcceptJob)
	r.With(middleware.RequireRoles("admin", "consumer")).Post("/api/v1/jobs/{id}/send-offer", api.SendJobOffer)

	// Job Workflow endpoints
	r.With(middleware.RequireRole("gig_worker")).Post("/api/v1/jobs/{id}/start", api.StartJob)
	r.With(middleware.RequireRole("gig_worker")).Post("/api/v1/jobs/{id}/complete", api.CompleteJob)
	r.With(middleware.RequireRole("gig_worker")).Post("/api/v1/jobs/{id}/reject", api.RejectJob)
	r.With(middleware.RequireRoles("admin", "consumer")).Post("/api/v1/jobs/{id}/review", api.SubmitReview)

	// Review Management
	r.With(middleware.RequireRoles("admin", "consumer", "gig_worker")).Post("/api/v1/reviews", api.CreateReview)

	// Schedule Management
	r.Post("/api/v1/schedules/create", api.CreateSchedule) // Any authenticated user

	// Transaction Management
	r.With(middleware.RequireRole("admin")).Post("/api/v1/transactions/create", api.CreateTransaction)
}

func PutHandlers(r chi.Router) {
	// User Management - Protected endpoints
	r.Put("/api/v1/users/profile", api.UpdateUserProfile) // Any authenticated user can update their own profile
	r.With(middleware.RequireRole("admin")).Put("/api/v1/users/{id}", api.UpdateUser)

	// GigWorker Management
	r.Put("/api/v1/gigworkers/{id}", api.UpdateGigWorker) // Any authenticated user (should validate ownership in handler)

	// Job Management
	r.With(middleware.RequireRoles("admin", "consumer")).Put("/api/v1/jobs/{id}", api.UpdateJob)

	// Review Management
	r.With(middleware.RequireRoles("admin", "consumer", "gig_worker")).Put("/api/v1/reviews/{id}", api.UpdateReview)
}

func DeleteHandlers(r chi.Router) {
	// User Management - Admin only
	r.With(middleware.RequireRole("admin")).Delete("/api/v1/users/{id}", api.DeactivateUser)

	// GigWorker Management - Admin only
	r.With(middleware.RequireRole("admin")).Delete("/api/v1/gigworkers/{id}", api.DeactivateGigWorker)

	// Job Management
	r.With(middleware.RequireRoles("admin", "consumer")).Delete("/api/v1/jobs/{id}", api.CancelJob)

	// Review Management
	r.With(middleware.RequireRoles("admin", "consumer", "gig_worker")).Delete("/api/v1/reviews/{id}", api.DeleteReview)
}
