package dnsimple

// User represents a DNSimple user.
type User struct {
	ID    int64  `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
}
