package externalaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/types"
)

type compiledHTTPResponse struct {
	statuses        types.HTTPCodeRanges
	successJSONPath string
	successValues   map[string]struct{}
	bodyRegex       *regexp.Regexp
	messageJSONPath string
}

func compileHTTPResponse(config *dynamic.ExternalActionHTTPResponse) (*compiledHTTPResponse, error) {
	if config == nil {
		return nil, nil
	}
	compiled := &compiledHTTPResponse{
		successJSONPath: strings.TrimSpace(config.SuccessJSONPath),
		successValues:   make(map[string]struct{}, len(config.SuccessValues)),
		messageJSONPath: strings.TrimSpace(config.MessageJSONPath),
	}
	if len(config.Statuses) > 0 {
		statuses, err := types.NewHTTPCodeRanges(config.Statuses)
		if err != nil {
			return nil, fmt.Errorf("parsing HTTP response validation statuses: %w", err)
		}
		compiled.statuses = statuses
	}
	for _, value := range config.SuccessValues {
		compiled.successValues[value] = struct{}{}
	}
	if config.BodyRegex != "" {
		expression, err := regexp.Compile(config.BodyRegex)
		if err != nil {
			return nil, fmt.Errorf("compiling HTTP response validation bodyRegex: %w", err)
		}
		compiled.bodyRegex = expression
	}
	return compiled, nil
}

func (r *compiledHTTPResponse) validate(body []byte) error {
	if r.bodyRegex != nil && !r.bodyRegex.Match(body) {
		message := r.errorMessage(body)
		if message != "" {
			return fmt.Errorf("external action response body did not match configured success pattern: %s", message)
		}
		return errors.New("external action response body did not match configured success pattern")
	}
	if r.successJSONPath == "" {
		return nil
	}

	payload, err := decodeJSON(body)
	if err != nil {
		return fmt.Errorf("external action response is not valid JSON: %w", err)
	}
	actualValue := selectJSON(payload, r.successJSONPath)
	actual := responseValueString(actualValue)
	if _, ok := r.successValues[actual]; ok {
		return nil
	}

	message := responseMessage(payload, r.messageJSONPath)
	if message != "" {
		return fmt.Errorf("external action response validation failed at %s (value %q): %s", r.successJSONPath, actual, message)
	}
	return fmt.Errorf("external action response validation failed at %s (value %q)", r.successJSONPath, actual)
}

func (r *compiledHTTPResponse) errorMessage(body []byte) string {
	if r.messageJSONPath != "" {
		payload, err := decodeJSON(body)
		if err == nil {
			if message := responseMessage(payload, r.messageJSONPath); message != "" {
				return message
			}
		}
	}
	return sanitizeResponseMessage(string(body))
}

func decodeJSON(body []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("multiple JSON values")
		}
		return nil, err
	}
	return payload, nil
}

func responseMessage(payload any, path string) string {
	if path == "" {
		return ""
	}
	value := selectJSON(payload, path)
	if value == nil {
		return ""
	}
	return sanitizeResponseMessage(responseValueString(value))
}

func responseValueString(value any) string {
	switch typed := value.(type) {
	case nil:
		return "<missing>"
	case string:
		return typed
	case json.Number:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(encoded)
	}
}

func sanitizeResponseMessage(value string) string {
	value = strings.Map(func(character rune) rune {
		if character < 0x20 || character == 0x7f {
			return ' '
		}
		return character
	}, strings.TrimSpace(value))
	const maxMessageRunes = 1024
	runes := []rune(value)
	if len(runes) > maxMessageRunes {
		return string(runes[:maxMessageRunes]) + "..."
	}
	return value
}
