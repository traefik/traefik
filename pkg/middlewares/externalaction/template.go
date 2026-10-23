package externalaction

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type compiledTemplate struct {
	value     *template.Template
	literal   string
	isLiteral bool
}

func newTemplate(name, value string) (*compiledTemplate, error) {
	functions := sprig.TxtFuncMap()
	delete(functions, "env")
	delete(functions, "expandenv")
	functions["jsonPath"] = selectJSON
	functions["header"] = func(data *actionData, name string) string {
		return data.Request.Headers.Get(name)
	}
	functions["query"] = func(data *actionData, name string) string {
		return data.Request.Query.Get(name)
	}
	// The signed string and header names belong to configuration, not a vendor.
	functions["hmacSHA256Base64"] = func(key, message string) string {
		mac := hmac.New(sha256.New, []byte(key))
		_, _ = mac.Write([]byte(message))
		return base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}
	functions["cookie"] = func(data *actionData, name string) string {
		return data.Request.Cookies[name]
	}
	// authParameter exposes the already configured per-target authentication
	// parameters to URL/Header/Body templates. This keeps project metadata in
	// the same rule-local source of truth as the outbound authentication
	// headers instead of hard-coding it in a business payload template.
	functions["authParameter"] = func(data *actionData, name string) string {
		return data.AuthParameters[name]
	}
	functions["parseInt"] = func(value any) (int64, error) {
		return strconv.ParseInt(strings.TrimSpace(fmt.Sprint(value)), 10, 64)
	}
	functions["scalarString"] = func(value any) (string, error) {
		switch typed := value.(type) {
		case string:
			return typed, nil
		case json.Number:
			return typed.String(), nil
		case float64:
			return strconv.FormatFloat(typed, 'f', -1, 64), nil
		case float32:
			return strconv.FormatFloat(float64(typed), 'f', -1, 32), nil
		case int:
			return strconv.Itoa(typed), nil
		case int8, int16, int32, int64:
			return fmt.Sprintf("%d", typed), nil
		case uint, uint8, uint16, uint32, uint64:
			return fmt.Sprintf("%d", typed), nil
		default:
			return "", fmt.Errorf("value of type %T is not a scalar string or number", value)
		}
	}
	functions["required"] = func(message string, value any) (any, error) {
		if isEmpty(value) {
			return nil, errors.New(message)
		}
		return value, nil
	}

	parsed, err := template.New(name).Option("missingkey=error").Funcs(functions).Parse(value)
	if err != nil {
		return nil, err
	}
	return &compiledTemplate{value: parsed}, nil
}

func (t *compiledTemplate) render(data *actionData) (string, error) {
	if t == nil {
		return "", nil
	}
	if t.isLiteral {
		return t.literal, nil
	}
	var buffer bytes.Buffer
	if err := t.value.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func newLiteralTemplate(value string) *compiledTemplate {
	return &compiledTemplate{literal: value, isLiteral: true}
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return reflected.Len() == 0
	case reflect.Bool:
		return !reflected.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return reflected.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return reflected.Float() == 0
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Pointer:
		return reflected.IsNil()
	default:
		return false
	}
}

func compileTemplates(prefix string, values map[string]string) (map[string]*compiledTemplate, error) {
	result := make(map[string]*compiledTemplate, len(values))
	for name, value := range values {
		parsed, err := newTemplate(prefix+"-"+name, value)
		if err != nil {
			return nil, fmt.Errorf("compiling %s %q: %w", prefix, name, err)
		}
		result[name] = parsed
	}
	return result, nil
}
