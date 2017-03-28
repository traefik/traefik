package dnsimple

// User represents a DNSimple user.
type User struct {
	ID    int    `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
}
