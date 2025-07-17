package handler

import (
	"app/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func GetHandlers(r chi.Router) {
	r.Get("/", middleware.ServeEmailForm)
	r.Get("/email-submit", middleware.HandleEmailSubmission)
}
