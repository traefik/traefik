package account

// Setting represents an accounts' contact info.
type Setting struct {
	CustomerID int     `json:"customerid,omitempty"`
	FirstName  string  `json:"firstname,omitempty"`
	LastName   string  `json:"lastname,omitempty"`
	Company    string  `json:"company,omitempty"`
	Phone      string  `json:"phone,omitempty"`
	Email      string  `json:"email,omitempty"`
	Address    Address `json:"address,omitempty"`
}

// Address for Setting struct.
type Address struct {
	Country string `json:"country,omitempty"`
	Street  string `json:"street,omitempty"`
	State   string `json:"state,omitempty"`
	City    string `json:"city,omitempty"`
	Postal  string `json:"postalcode,omitempty"`
}
