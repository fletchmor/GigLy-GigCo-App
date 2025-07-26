<<<<<<< HEAD
package model

type User struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	ID      int    `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
=======
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
>>>>>>> 86c55890c09c3f69b573f338c0a66df1e5fdb519
