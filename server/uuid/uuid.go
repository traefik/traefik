package uuid

import guuid "github.com/google/uuid"

var uuid string

func init() {
	uuid = guuid.New().String()
}

// Get the instance UUID
func Get() string {
	return uuid
}
