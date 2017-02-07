package account

// APIKey wraps an NS1 /account/apikeys resource
type APIKey struct {
	// Read-only fields
	ID         string `json:"id,omitempty"`
	Key        string `json:"key,omitempty"`
	LastAccess int    `json:"last_access,omitempty"`

	Name        string         `json:"name"`
	TeamIDs     []string       `json:"teams"`
	Permissions PermissionsMap `json:"permissions"`
}
