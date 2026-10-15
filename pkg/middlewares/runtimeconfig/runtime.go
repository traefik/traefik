// Package runtimeconfig resolves arbitrarily named runtime values without
// publishing their material in Traefik's dynamic configuration APIs.
package runtimeconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

const referencePrefix = "runtime://"

var (
	aliasPattern         = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,127}$`)
	interpolationPattern = regexp.MustCompile(`\$\{runtime\.([A-Za-z][A-Za-z0-9_.-]{0,127})\}`)
	runtimeValueType     = reflect.TypeFor[dynamic.RuntimeValue]()
)

// Resolve replaces runtime references in every string field of target.
//
// An exact reference uses runtime://alias. Embedded values use
// ${runtime.alias}. RuntimeValues themselves are never traversed.
func Resolve(values map[string]dynamic.RuntimeValue, target any) error {
	if target == nil {
		return errors.New("runtime target is nil")
	}

	loaded, err := load(values)
	if err != nil {
		return err
	}

	root := reflect.ValueOf(target)
	if root.Kind() != reflect.Pointer || root.IsNil() {
		return errors.New("runtime target must be a non-nil pointer")
	}
	return resolveValue(root.Elem(), loaded)
}

// SourceIdentity returns a stable, non-secret identity for a connection value.
// The identity remains stable when a Secret or file's contents rotate.
func SourceIdentity(value, file string, values map[string]dynamic.RuntimeValue) (string, error) {
	value = strings.TrimSpace(value)
	file = strings.TrimSpace(file)
	if value != "" && file != "" {
		return "", errors.New("connection must not define both dsn and dsnFile")
	}
	if file != "" {
		return "file:" + file, nil
	}
	if alias, ok := ExactAlias(value); ok {
		source, exists := values[alias]
		if !exists {
			return "", fmt.Errorf("runtime value %q is not declared", alias)
		}
		if strings.TrimSpace(source.Source) != "" {
			return "runtime:" + source.Source, nil
		}
		if strings.TrimSpace(source.File) != "" {
			return "runtime:file:" + source.File, nil
		}
		return "runtime:literal:" + alias, nil
	}
	if value == "" {
		return "", errors.New("connection requires dsn or dsnFile")
	}
	sum := sha256.Sum256([]byte(value))
	return "literal:" + hex.EncodeToString(sum[:]), nil
}

// ExactAlias parses an exact runtime://alias reference.
func ExactAlias(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, referencePrefix) {
		return "", false
	}
	alias := strings.TrimSpace(strings.TrimPrefix(value, referencePrefix))
	return alias, aliasPattern.MatchString(alias)
}

// ExactReferenceFile returns the file backing an exact runtime reference.
func ExactReferenceFile(value string, values map[string]dynamic.RuntimeValue) string {
	alias, ok := ExactAlias(value)
	if !ok {
		return ""
	}
	return strings.TrimSpace(values[alias].File)
}

func load(values map[string]dynamic.RuntimeValue) (map[string]string, error) {
	loaded := make(map[string]string, len(values))
	for alias, source := range values {
		if !aliasPattern.MatchString(alias) {
			return nil, fmt.Errorf("runtime value name %q is invalid", alias)
		}
		if strings.TrimSpace(source.File) != "" {
			data, err := os.ReadFile(source.File)
			if err != nil {
				return nil, fmt.Errorf("reading runtime value %q file: %w", alias, err)
			}
			loaded[alias] = strings.TrimSpace(string(data))
			continue
		}
		loaded[alias] = source.Value
	}
	return loaded, nil
}

func resolveString(value string, values map[string]string) (string, error) {
	if alias, ok := ExactAlias(value); ok {
		resolved, exists := values[alias]
		if !exists {
			return "", fmt.Errorf("runtime value %q is not declared", alias)
		}
		return resolved, nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), referencePrefix) {
		return "", fmt.Errorf("invalid runtime reference %q", value)
	}

	var missing string
	result := interpolationPattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := interpolationPattern.FindStringSubmatch(match)
		resolved, exists := values[parts[1]]
		if !exists {
			missing = parts[1]
			return match
		}
		return resolved
	})
	if missing != "" {
		return "", fmt.Errorf("runtime value %q is not declared", missing)
	}
	return result, nil
}

func resolveValue(value reflect.Value, values map[string]string) error {
	if !value.IsValid() {
		return nil
	}
	if value.Type() == runtimeValueType {
		return nil
	}

	switch value.Kind() {
	case reflect.Pointer:
		if !value.IsNil() {
			return resolveValue(value.Elem(), values)
		}
	case reflect.Interface:
		if value.IsNil() {
			return nil
		}
		copyValue := reflect.New(value.Elem().Type()).Elem()
		copyValue.Set(value.Elem())
		if err := resolveValue(copyValue, values); err != nil {
			return err
		}
		value.Set(copyValue)
	case reflect.Struct:
		for _, field := range value.Fields() {
			if field.CanSet() {
				if err := resolveValue(field, values); err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		for index := range value.Len() {
			if err := resolveValue(value.Index(index), values); err != nil {
				return err
			}
		}
	case reflect.Array:
		for index := range value.Len() {
			if err := resolveValue(value.Index(index), values); err != nil {
				return err
			}
		}
	case reflect.Map:
		if value.IsNil() {
			return nil
		}
		iterator := value.MapRange()
		for iterator.Next() {
			key := iterator.Key()
			entry := iterator.Value()
			copyValue := reflect.New(entry.Type()).Elem()
			copyValue.Set(entry)
			if err := resolveValue(copyValue, values); err != nil {
				return err
			}
			value.SetMapIndex(key, copyValue)
		}
	case reflect.String:
		resolved, err := resolveString(value.String(), values)
		if err != nil {
			return err
		}
		value.SetString(resolved)
	}
	return nil
}
