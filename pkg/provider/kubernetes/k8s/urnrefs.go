package k8s

import (
	"strconv"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// SecretURNPrefix is the prefix of the URN referencing the value of a key of a Secret.
const SecretURNPrefix = "urn:k8s:secret:"

// SecretURN returns the URN referencing the value of the given key of the given Secret.
func SecretURN(secretName, key string) string {
	return SecretURNPrefix + secretName + ":" + key
}

// URNRefs collects the values read from Secrets for a middleware, keyed by the JSON Pointer of the field referencing them.
type URNRefs map[string]string

// Add records the value under the JSON Pointer, and returns the URN of the Secret key to set in the field.
func (r URNRefs) Add(pointer, secretName, key, value string) string {
	r[pointer] = value

	return SecretURN(secretName, key)
}

// AddUsers records each user under the JSON Pointer of its entry in the list at pointer, and returns one URN per user.
func (r URNRefs) AddUsers(pointer, secretName, key string, users []string) dynamic.Users {
	urns := make(dynamic.Users, len(users))
	for i, user := range users {
		urns[i] = r.Add(pointer+"/"+strconv.Itoa(i), secretName, key, user)
	}

	return urns
}

// OrNil returns nil when no value was recorded, so middlewares without Secret keep the configuration they had.
func (r URNRefs) OrNil() map[string]string {
	if len(r) == 0 {
		return nil
	}

	return r
}
