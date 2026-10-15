package crd

import (
	"strconv"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
)

// urnRefs collects the values read from Secrets for a middleware, keyed by the JSON Pointer of the field referencing them.
type urnRefs map[string]string

// add records the value under the JSON Pointer, and returns the URN of the Secret key to set in the field.
func (r urnRefs) add(pointer, secretName, key, value string) string {
	r[pointer] = value

	return k8s.SecretURN(secretName, key)
}

// addUsers records each user under the JSON Pointer of its entry in the list at pointer, and returns one URN per user.
func (r urnRefs) addUsers(pointer, secretName, key string, users []string) dynamic.Users {
	urns := make(dynamic.Users, len(users))
	for i, user := range users {
		urns[i] = r.add(pointer+"/"+strconv.Itoa(i), secretName, key, user)
	}

	return urns
}

// orNil returns nil when no value was recorded, so middlewares without Secret keep the configuration they had.
func (r urnRefs) orNil() map[string]string {
	if len(r) == 0 {
		return nil
	}

	return r
}
