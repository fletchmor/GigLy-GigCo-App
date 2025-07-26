package handler

import (
	"app/api"
	"app/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func GetHandlers(r chi.Router) {
	r.Get("/", middleware.ServeEmailForm)
	r.Get("/email-submit", middleware.HandleEmailSubmission)
	r.Post("/api/v1/users/create", api.CreateUser)
	r.Get("/api/v1/customers/{id}", api.GetCustomerByID)
}
