package labels

import (
	"encoding/json"
	"fmt"

	"github.com/docker/libcompose/utils"
)

// Label represents a docker label.
type Label string

// Libcompose default labels.
const (
	NUMBER  = Label("com.docker.compose.container-number")
	ONEOFF  = Label("com.docker.compose.oneoff")
	PROJECT = Label("com.docker.compose.project")
	SERVICE = Label("com.docker.compose.service")
	HASH    = Label("com.docker.compose.config-hash")
	VERSION = Label("com.docker.compose.version")
)

// EqString returns a label json string representation with the specified value.
func (f Label) EqString(value string) string {
	return LabelFilterString(string(f), value)
}

// Eq returns a label map representation with the specified value.
func (f Label) Eq(value string) map[string][]string {
	return LabelFilter(string(f), value)
}

// AndString returns a json list of labels by merging the two specified values (left and right) serialized as string.
func AndString(left, right string) string {
	leftMap := map[string][]string{}
	rightMap := map[string][]string{}

	// Ignore errors
	json.Unmarshal([]byte(left), &leftMap)
	json.Unmarshal([]byte(right), &rightMap)

	for k, v := range rightMap {
		existing, ok := leftMap[k]
		if ok {
			leftMap[k] = append(existing, v...)
		} else {
			leftMap[k] = v
		}
	}

	result, _ := json.Marshal(leftMap)

	return string(result)
}

// And returns a map of labels by merging the two specified values (left and right).
func And(left, right map[string][]string) map[string][]string {
	result := map[string][]string{}
	for k, v := range left {
		result[k] = v
	}

	for k, v := range right {
		existing, ok := result[k]
		if ok {
			result[k] = append(existing, v...)
		} else {
			result[k] = v
		}
	}

	return result
}

// Str returns the label name.
func (f Label) Str() string {
	return string(f)
}

// LabelFilterString returns a label json string representation of the specifed couple (key,value)
// that is used as filter for docker.
func LabelFilterString(key, value string) string {
	return utils.FilterString(map[string][]string{
		"label": {fmt.Sprintf("%s=%s", key, value)},
	})
}

// LabelFilter returns a label map representation of the specifed couple (key,value)
// that is used as filter for docker.
func LabelFilter(key, value string) map[string][]string {
	return map[string][]string{
		"label": {fmt.Sprintf("%s=%s", key, value)},
	}
}
