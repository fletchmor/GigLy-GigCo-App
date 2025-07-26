package model

type User struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	ID      int    `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
