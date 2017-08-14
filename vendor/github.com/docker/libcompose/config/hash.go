package config

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/docker/libcompose/yaml"
)

// GetServiceHash computes and returns a hash that will identify a service.
// This hash will be then used to detect if the service definition/configuration
// have changed and needs to be recreated.
func GetServiceHash(name string, config *ServiceConfig) string {
	hash := sha1.New()

	io.WriteString(hash, name)

	//Get values of Service through reflection
	val := reflect.ValueOf(config).Elem()

	//Create slice to sort the keys in Service Config, which allow constant hash ordering
	serviceKeys := []string{}

	//Create a data structure of map of values keyed by a string
	unsortedKeyValue := make(map[string]interface{})

	//Get all keys and values in Service Configuration
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		keyField := val.Type().Field(i)

		serviceKeys = append(serviceKeys, keyField.Name)
		unsortedKeyValue[keyField.Name] = valueField.Interface()
	}

	//Sort serviceKeys alphabetically
	sort.Strings(serviceKeys)

	//Go through keys and write hash
	for _, serviceKey := range serviceKeys {
		serviceValue := unsortedKeyValue[serviceKey]

		io.WriteString(hash, fmt.Sprintf("\n  %v: ", serviceKey))

		switch s := serviceValue.(type) {
		case yaml.SliceorMap:
			sliceKeys := []string{}
			for lkey := range s {
				sliceKeys = append(sliceKeys, lkey)
			}
			sort.Strings(sliceKeys)

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s=%v, ", sliceKey, s[sliceKey]))
			}
		case yaml.MaporEqualSlice:
			for _, sliceKey := range s {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case yaml.MaporColonSlice:
			for _, sliceKey := range s {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case yaml.MaporSpaceSlice:
			for _, sliceKey := range s {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case yaml.Command:
			for _, sliceKey := range s {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case yaml.Stringorslice:
			sort.Strings(s)

			for _, sliceKey := range s {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case []string:
			sliceKeys := s
			sort.Strings(sliceKeys)

			for _, sliceKey := range sliceKeys {
				io.WriteString(hash, fmt.Sprintf("%s, ", sliceKey))
			}
		case *yaml.Networks:
			io.WriteString(hash, fmt.Sprintf("%s, ", s.HashString()))
		case *yaml.Volumes:
			io.WriteString(hash, fmt.Sprintf("%s, ", s.HashString()))
		default:
			io.WriteString(hash, fmt.Sprintf("%v, ", serviceValue))
		}
	}

	return hex.EncodeToString(hash.Sum(nil))
}
