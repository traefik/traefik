package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type BasicAuth struct {
	Username string
	Password string
}

func (ba *BasicAuth) String() string {
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", ba.Username, ba.Password)))
	return fmt.Sprintf("Basic %s", encoded)
}

func ParseAuthHeader(header string) (*BasicAuth, error) {
	values := strings.Fields(header)
	if len(values) != 2 {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to parse header '%s'", header))
	}

	auth_type := strings.ToLower(values[0])
	if auth_type != "basic" {
		return nil, fmt.Errorf("Expected basic auth type, got '%s'", auth_type)
	}

	encoded_string := values[1]
	decoded_string, err := base64.StdEncoding.DecodeString(encoded_string)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse header '%s', base64 failed: %s", header, err)
	}

	values = strings.SplitN(string(decoded_string), ":", 2)
	if len(values) != 2 {
		return nil, fmt.Errorf("Failed to parse header '%s', expected separator ':'", header)
	}
	return &BasicAuth{Username: values[0], Password: values[1]}, nil
}
