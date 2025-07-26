package model

import "time"

type User struct {
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
