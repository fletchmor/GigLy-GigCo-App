package model

type Address struct {
	StreetName string `json:"streetname"`
	State      string `json:"state"`
	PostalCode int    `json:"postalcode"`
}

type Name struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type User struct {
	Address Address `json:"address"`
	Name    Name    `json:"name"`
	Email   string  `json:"email"`
}
