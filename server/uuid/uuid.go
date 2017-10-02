package uuid

import guuid "github.com/satori/go.uuid"

var uuid string

func init() {
	uuid = guuid.NewV4().String()
}

// Get the instance UUID
func Get() string {
	return uuid
}
