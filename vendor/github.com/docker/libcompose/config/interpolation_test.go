package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func testInterpolatedLine(t *testing.T, expectedLine, interpolatedLine string, envVariables map[string]string) {
	interpolatedLine, _ = parseLine(interpolatedLine, func(s string) string {
		return envVariables[s]
	})

	assert.Equal(t, expectedLine, interpolatedLine)
}

func testInvalidInterpolatedLine(t *testing.T, line string) {
	_, success := parseLine(line, func(string) string {
		return ""
	})

	assert.Equal(t, false, success)
}

func TestParseLine(t *testing.T) {
	variables := map[string]string{
		"A":           "ABC",
		"X":           "XYZ",
		"E":           "",
		"lower":       "WORKED",
		"MiXeD":       "WORKED",
		"split_VaLue": "WORKED",
		"9aNumber":    "WORKED",
		"a9Number":    "WORKED",
	}

	testInterpolatedLine(t, "WORKED", "$lower", variables)
	testInterpolatedLine(t, "WORKED", "${MiXeD}", variables)
	testInterpolatedLine(t, "WORKED", "${split_VaLue}", variables)
	// Starting with a number isn't valid
	testInterpolatedLine(t, "", "$9aNumber", variables)
	testInterpolatedLine(t, "WORKED", "$a9Number", variables)

	testInterpolatedLine(t, "ABC", "$A", variables)
	testInterpolatedLine(t, "ABC", "${A}", variables)

	testInterpolatedLine(t, "ABC DE", "$A DE", variables)
	testInterpolatedLine(t, "ABCDE", "${A}DE", variables)

	testInterpolatedLine(t, "$A", "$$A", variables)
	testInterpolatedLine(t, "${A}", "$${A}", variables)

	testInterpolatedLine(t, "$ABC", "$$${A}", variables)
	testInterpolatedLine(t, "$ABC", "$$$A", variables)

	testInterpolatedLine(t, "ABC XYZ", "$A $X", variables)
	testInterpolatedLine(t, "ABCXYZ", "$A$X", variables)
	testInterpolatedLine(t, "ABCXYZ", "${A}${X}", variables)

	testInterpolatedLine(t, "", "$B", variables)
	testInterpolatedLine(t, "", "${B}", variables)
	testInterpolatedLine(t, "", "$ADE", variables)

	testInterpolatedLine(t, "", "$E", variables)
	testInterpolatedLine(t, "", "${E}", variables)

	testInvalidInterpolatedLine(t, "${")
	testInvalidInterpolatedLine(t, "$}")
	testInvalidInterpolatedLine(t, "${}")
	testInvalidInterpolatedLine(t, "${ }")
	testInvalidInterpolatedLine(t, "${A }")
	testInvalidInterpolatedLine(t, "${ A}")
	testInvalidInterpolatedLine(t, "${A!}")
	testInvalidInterpolatedLine(t, "$!")
}

type MockEnvironmentLookup struct {
	Variables map[string]string
}

func (m MockEnvironmentLookup) Lookup(key string, config *ServiceConfig) []string {
	return []string{fmt.Sprintf("%s=%s", key, m.Variables[key])}
}

func testInterpolatedConfig(t *testing.T, expectedConfig, interpolatedConfig string, envVariables map[string]string) {
	for k, v := range envVariables {
		os.Setenv(k, v)
	}

	expectedConfigBytes := []byte(expectedConfig)
	interpolatedConfigBytes := []byte(interpolatedConfig)

	expectedData := make(RawServiceMap)
	interpolatedData := make(RawServiceMap)

	yaml.Unmarshal(expectedConfigBytes, &expectedData)
	yaml.Unmarshal(interpolatedConfigBytes, &interpolatedData)

	_ = InterpolateRawServiceMap(&interpolatedData, MockEnvironmentLookup{envVariables})

	for k := range envVariables {
		os.Unsetenv(k)
	}

	assert.Equal(t, expectedData, interpolatedData)
}

func testInvalidInterpolatedConfig(t *testing.T, interpolatedConfig string) {
	interpolatedConfigBytes := []byte(interpolatedConfig)
	interpolatedData := make(RawServiceMap)
	yaml.Unmarshal(interpolatedConfigBytes, &interpolatedData)

	err := InterpolateRawServiceMap(&interpolatedData, new(MockEnvironmentLookup))
	assert.NotNil(t, err)
}

func TestInterpolate(t *testing.T) {
	testInterpolatedConfig(t,
		`web:
  # unbracketed name
  image: busybox

  # array element
  ports:
    - "80:8000"

  # dictionary item value
  labels:
    mylabel: "myvalue"

  # unset value
  hostname: "host-"

  # escaped interpolation
  command: "${ESCAPED}"`,
		`web:
  # unbracketed name
  image: $IMAGE

  # array element
  ports:
    - "${HOST_PORT}:8000"

  # dictionary item value
  labels:
    mylabel: "${LABEL_VALUE}"

  # unset value
  hostname: "host-${UNSET_VALUE}"

  # escaped interpolation
  command: "$${ESCAPED}"`, map[string]string{
			"IMAGE":       "busybox",
			"HOST_PORT":   "80",
			"LABEL_VALUE": "myvalue",
		})

	// Same as above, but testing with equal signs in variables
	testInterpolatedConfig(t,
		`web:
  # unbracketed name
  image: =busybox

  # array element
  ports:
    - "=:8000"

  # dictionary item value
  labels:
    mylabel: "myvalue=="

  # unset value
  hostname: "host-"

  # escaped interpolation
  command: "${ESCAPED}"`,
		`web:
  # unbracketed name
  image: $IMAGE

  # array element
  ports:
    - "${HOST_PORT}:8000"

  # dictionary item value
  labels:
    mylabel: "${LABEL_VALUE}"

  # unset value
  hostname: "host-${UNSET_VALUE}"

  # escaped interpolation
  command: "$${ESCAPED}"`, map[string]string{
			"IMAGE":       "=busybox",
			"HOST_PORT":   "=",
			"LABEL_VALUE": "myvalue==",
		})

	testInvalidInterpolatedConfig(t,
		`web:
  image: "${"`)

	testInvalidInterpolatedConfig(t,
		`web:
  image: busybox

  # array element
  ports:
    - "${}:8000"`)

	testInvalidInterpolatedConfig(t,
		`web:
  image: busybox

  # array element
  ports:
    - "80:8000"

  # dictionary item value
  labels:
    mylabel: "${ LABEL_VALUE}"`)
}
