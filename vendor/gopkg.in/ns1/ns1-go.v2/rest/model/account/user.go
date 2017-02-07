package account

// User wraps an NS1 /account/users resource
type User struct {
	// Read-only fields
	LastAccess float64 `json:"last_access"`

	Name        string               `json:"name"`
	Username    string               `json:"username"`
	Email       string               `json:"email"`
	TeamIDs     []string             `json:"teams"`
	Notify      NotificationSettings `json:"notify"`
	Permissions PermissionsMap       `json:"permissions"`
}

// NotificationSettings wraps a User's "notify" attribute
type NotificationSettings struct {
	Billing bool `json:"billing"`
}
