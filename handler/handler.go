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
	r.Get("/api/v1/customers/{id}", api.GetCustomerByID)
	r.Get("/api/v1/jobs", api.GetJobs)
	r.Get("/api/v1/jobs/{id}", api.GetJobByID)
}

func PostHandlers(r chi.Router) {
	r.Post("/api/v1/users/create", api.CreateUser)
	r.Post("/api/v1/schedules/create", api.CreateSchedule)
	r.Post("/api/v1/transactions/create", api.CreateTransaction)
	r.Post("/api/v1/auth/register", api.RegisterUser)
	r.Post("/api/v1/jobs/create", api.CreateJob)
	r.Post("/api/v1/jobs/{id}/accept", api.AcceptJob)
}
