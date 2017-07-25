package config

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testValidSchemaV1(t *testing.T, serviceMap RawServiceMap) {
	testValidSchema(t, serviceMap, validate, validateServiceConstraints)
}

func testValidSchemaV2(t *testing.T, serviceMap RawServiceMap) {
	testValidSchema(t, serviceMap, validateV2, nil)
}

func testValidSchemaAll(t *testing.T, serviceMap RawServiceMap) {
	testValidSchema(t, serviceMap, validate, validateServiceConstraints)
	testValidSchema(t, serviceMap, validateV2, nil)
}

func testValidSchema(t *testing.T, serviceMap RawServiceMap, validate func(RawServiceMap) error, validateServiceConstraints func(RawService, string) error) {
	err := validate(serviceMap)
	assert.Nil(t, err)

	if validateServiceConstraints != nil {
		for name, service := range serviceMap {
			err := validateServiceConstraints(service, name)
			assert.Nil(t, err)
		}
	}
}

func testInvalidSchemaV1(t *testing.T, serviceMap RawServiceMap, errMsgs []string, errCount int) {
	testInvalidSchema(t, serviceMap, errMsgs, errCount, validate, validateServiceConstraints)
}

func testInvalidSchemaV2(t *testing.T, serviceMap RawServiceMap, errMsgs []string, errCount int) {
	testInvalidSchema(t, serviceMap, errMsgs, errCount, validateV2, nil)
}

func testInvalidSchemaAll(t *testing.T, serviceMap RawServiceMap, errMsgs []string, errCount int) {
	testInvalidSchema(t, serviceMap, errMsgs, errCount, validate, validateServiceConstraints)
	testInvalidSchema(t, serviceMap, errMsgs, errCount, validateV2, nil)
}

func testInvalidSchema(t *testing.T, serviceMap RawServiceMap, errMsgs []string, errCount int, validate func(RawServiceMap) error, validateServiceConstraints func(RawService, string) error) {
	var combinedErrMsg bytes.Buffer

	err := validate(serviceMap)
	if err != nil {
		combinedErrMsg.WriteString(err.Error())
		combinedErrMsg.WriteRune('\n')
	}

	if validateServiceConstraints != nil {
		for name, service := range serviceMap {
			err := validateServiceConstraints(service, name)
			if err != nil {
				combinedErrMsg.WriteString(err.Error())
				combinedErrMsg.WriteRune('\n')
			}
		}
	}

	for _, errMsg := range errMsgs {
		assert.True(t, strings.Contains(combinedErrMsg.String(), errMsg))
	}

	// gojsonschema has bugs that can cause extraneous errors
	// This makes sure we don't have more errors than expected
	assert.True(t, strings.Count(combinedErrMsg.String(), "\n") == errCount)
}

func TestInvalidServiceNames(t *testing.T) {
	invalidServiceNames := []string{"?not?allowed", " ", "", "!", "/"}

	for _, invalidServiceName := range invalidServiceNames {
		testInvalidSchemaAll(t, RawServiceMap{
			invalidServiceName: map[string]interface{}{
				"image": "busybox",
			},
		}, []string{fmt.Sprintf("Invalid service name '%s' - only [a-zA-Z0-9\\._\\-] characters are allowed", invalidServiceName)}, 1)
	}
}

func TestValidServiceNames(t *testing.T) {
	validServiceNames := []string{"_", "-", ".__.", "_what-up.", "what_.up----", "whatup"}

	for _, validServiceName := range validServiceNames {
		testValidSchemaAll(t, RawServiceMap{
			validServiceName: map[string]interface{}{
				"image": "busybox",
			},
		})
	}
}

func TestConfigInvalidPorts(t *testing.T) {
	portsValues := []interface{}{
		map[string]interface{}{
			"1": "8000",
		},
		false,
		0,
		"8000",
	}

	for _, portsValue := range portsValues {
		testInvalidSchemaAll(t, RawServiceMap{
			"web": map[string]interface{}{
				"image": "busybox",
				"ports": portsValue,
			},
		}, []string{"Service 'web' configuration key 'ports' contains an invalid type, it should be an array"}, 1)
	}

	testInvalidSchemaAll(t, RawServiceMap{
		"web": map[string]interface{}{
			"image": "busybox",
			"ports": []interface{}{
				"8000",
				"8000",
			},
		},
	}, []string{"Service 'web' configuration key 'ports' value [8000 8000] has non-unique elements"}, 1)
}

func TestConfigValidPorts(t *testing.T) {
	portsValues := []interface{}{
		[]interface{}{"8000", "9000"},
		[]interface{}{"8000"},
		[]interface{}{8000},
		[]interface{}{"127.0.0.1::8000"},
		[]interface{}{"49153-49154:3002-3003"},
	}

	for _, portsValue := range portsValues {
		testValidSchemaAll(t, RawServiceMap{
			"web": map[string]interface{}{
				"image": "busybox",
				"ports": portsValue,
			},
		})
	}
}

func TestConfigHint(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"foo": map[string]interface{}{
			"image":     "busybox",
			"privilege": "something",
		},
	}, []string{"Unsupported config option for foo service: 'privilege' (did you mean 'privileged'?)"}, 1)
}

func TestTypeShouldBeAnArray(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"foo": map[string]interface{}{
			"image": "busybox",
			"links": "an_link",
		},
	}, []string{"Service 'foo' configuration key 'links' contains an invalid type, it should be an array"}, 1)
}

func TestInvalidTypeWithMultipleValidTypes(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"web": map[string]interface{}{
			"image": "busybox",
			"mem_limit": []interface{}{
				"array_elem",
			},
		},
	}, []string{"Service 'web' configuration key 'mem_limit' contains an invalid type, it should be a number or string."}, 1)
}

func TestInvalidNotUniqueItems(t *testing.T) {
	// Test property with array as only valid type
	testInvalidSchemaAll(t, RawServiceMap{
		"foo": map[string]interface{}{
			"image": "busybox",
			"devices": []string{
				"/dev/foo:/dev/foo",
				"/dev/foo:/dev/foo",
			},
		},
	}, []string{"Service 'foo' configuration key 'devices' value [/dev/foo:/dev/foo /dev/foo:/dev/foo] has non-unique elements"}, 1)

	// Test property with multiple valid types
	testInvalidSchemaAll(t, RawServiceMap{
		"foo": map[string]interface{}{
			"image": "busybox",
			"environment": []string{
				"KEY=VAL",
				"KEY=VAL",
			},
		},
	}, []string{"Service 'foo' configuration key 'environment' contains non unique items, please remove duplicates from [KEY=VAL KEY=VAL]"}, 1)
}

func TestInvalidListOfStringsFormat(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"web": map[string]interface{}{
			"build": ".",
			"command": []interface{}{
				1,
			},
		},
	}, []string{"Service 'web' configuration key 'command' contains 1, which is an invalid type, it should be a string"}, 1)
}

func TestInvalidExtraHostsString(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"web": map[string]interface{}{
			"image":       "busybox",
			"extra_hosts": "somehost:162.242.195.82",
		},
	}, []string{"Service 'web' configuration key 'extra_hosts' contains an invalid type, it should be an array or object"}, 1)
}

func TestValidConfigWhichAllowsTwoTypeDefinitions(t *testing.T) {
	for _, exposeValue := range []interface{}{"8000", 9000} {
		testValidSchemaAll(t, RawServiceMap{
			"web": map[string]interface{}{
				"image": "busybox",
				"expose": []interface{}{
					exposeValue,
				},
			},
		})
	}
}

func TestValidConfigOneOfStringOrList(t *testing.T) {
	entrypointValues := []interface{}{
		[]interface{}{
			"sh",
		},
		"sh",
	}

	for _, entrypointValue := range entrypointValues {
		testValidSchemaAll(t, RawServiceMap{
			"web": map[string]interface{}{
				"image":      "busybox",
				"entrypoint": entrypointValue,
			},
		})
	}
}

func TestInvalidServiceProperty(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"web": map[string]interface{}{
			"image":            "busybox",
			"invalid_property": "value",
		},
	}, []string{"Unsupported config option for web service: 'invalid_property'"}, 1)
}

func TestServiceInvalidMissingImageAndBuild(t *testing.T) {
	testInvalidSchemaV1(t, RawServiceMap{
		"web": map[string]interface{}{},
	}, []string{"Service 'web' has neither an image nor a build path specified. Exactly one must be provided."}, 1)
}

func TestServiceInvalidSpecifiesImageAndBuild(t *testing.T) {
	testInvalidSchemaV1(t, RawServiceMap{
		"web": map[string]interface{}{
			"image": "busybox",
			"build": ".",
		},
	}, []string{"Service 'web' has both an image and build path specified. A service can either be built to image or use an existing image, not both."}, 1)
}

func TestServiceInvalidSpecifiesImageAndDockerfile(t *testing.T) {
	testInvalidSchemaV1(t, RawServiceMap{
		"web": map[string]interface{}{
			"image":      "busybox",
			"dockerfile": "Dockerfile",
		},
	}, []string{"Service 'web' has both an image and alternate Dockerfile. A service can either be built to image or use an existing image, not both."}, 1)
}

func TestInvalidServiceForMultipleErrors(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"foo": map[string]interface{}{
			"image": "busybox",
			"ports": "invalid_type",
			"links": "an_type",
			"environment": []string{
				"KEY=VAL",
				"KEY=VAL",
			},
		},
	}, []string{
		"Service 'foo' configuration key 'ports' contains an invalid type, it should be an array",
		"Service 'foo' configuration key 'links' contains an invalid type, it should be an array",
		"Service 'foo' configuration key 'environment' contains non unique items, please remove duplicates from [KEY=VAL KEY=VAL]",
	}, 3)
}

func TestInvalidServiceWithAdditionalProperties(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"foo": map[string]interface{}{
			"image": "busybox",
			"ports": "invalid_type",
			"---":   "nope",
			"environment": []string{
				"KEY=VAL",
				"KEY=VAL",
			},
		},
	}, []string{
		"Service 'foo' configuration key 'ports' contains an invalid type, it should be an array",
		"Unsupported config option for foo service: '---'",
		"Service 'foo' configuration key 'environment' contains non unique items, please remove duplicates from [KEY=VAL KEY=VAL]",
	}, 3)
}

func TestMultipleInvalidServices(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"foo1": map[string]interface{}{
			"image": "busybox",
			"ports": "invalid_type",
		},
		"foo2": map[string]interface{}{
			"image": "busybox",
			"ports": "invalid_type",
		},
	}, []string{
		"Service 'foo1' configuration key 'ports' contains an invalid type, it should be an array",
		"Service 'foo2' configuration key 'ports' contains an invalid type, it should be an array",
	}, 2)
}

func TestMixedInvalidServicesAndInvalidServiceNames(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"foo1": map[string]interface{}{
			"image": "busybox",
			"ports": "invalid_type",
		},
		"???": map[string]interface{}{
			"image": "busybox",
		},
		"foo2": map[string]interface{}{
			"image": "busybox",
			"ports": "invalid_type",
		},
	}, []string{
		"Service 'foo1' configuration key 'ports' contains an invalid type, it should be an array",
		"Invalid service name '???' - only [a-zA-Z0-9\\._\\-] characters are allowed",
		"Service 'foo2' configuration key 'ports' contains an invalid type, it should be an array",
	}, 3)
}

func TestMultipleInvalidServicesForMultipleErrors(t *testing.T) {
	testInvalidSchemaAll(t, RawServiceMap{
		"foo1": map[string]interface{}{
			"image": "busybox",
			"ports": "invalid_type",
			"environment": []string{
				"KEY=VAL",
				"KEY=VAL",
			},
		},
		"foo2": map[string]interface{}{
			"image": "busybox",
			"ports": "invalid_type",
			"environment": []string{
				"KEY=VAL",
				"KEY=VAL",
			},
		},
	}, []string{
		"Service 'foo1' configuration key 'ports' contains an invalid type, it should be an array",
		"Service 'foo1' configuration key 'environment' contains non unique items, please remove duplicates from [KEY=VAL KEY=VAL]",
		"Service 'foo2' configuration key 'ports' contains an invalid type, it should be an array",
		"Service 'foo2' configuration key 'environment' contains non unique items, please remove duplicates from [KEY=VAL KEY=VAL]",
	}, 4)
}
