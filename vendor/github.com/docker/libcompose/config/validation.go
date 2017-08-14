package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/libcompose/utils"
	"github.com/xeipuuv/gojsonschema"
)

func serviceNameFromErrorField(field string) string {
	splitKeys := strings.Split(field, ".")
	return splitKeys[0]
}

func keyNameFromErrorField(field string) string {
	splitKeys := strings.Split(field, ".")

	if len(splitKeys) > 0 {
		return splitKeys[len(splitKeys)-1]
	}

	return ""
}

func containsTypeError(resultError gojsonschema.ResultError) bool {
	contextSplit := strings.Split(resultError.Context().String(), ".")
	_, err := strconv.Atoi(contextSplit[len(contextSplit)-1])
	return err == nil
}

func addArticle(s string) string {
	switch s[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an " + s
	default:
		return "a " + s
	}
}

// Gets the value in a service map at a given error context
func getValue(val interface{}, context string) string {
	keys := strings.Split(context, ".")

	if keys[0] == "(root)" {
		keys = keys[1:]
	}

	for i, k := range keys {
		switch typedVal := (val).(type) {
		case string:
			return typedVal
		case []interface{}:
			if index, err := strconv.Atoi(k); err == nil {
				val = typedVal[index]
			}
		case RawServiceMap:
			val = typedVal[k]
		case RawService:
			val = typedVal[k]
		case map[interface{}]interface{}:
			val = typedVal[k]
		}

		if i == len(keys)-1 {
			return fmt.Sprint(val)
		}
	}

	return ""
}

func convertServiceMapKeysToStrings(serviceMap RawServiceMap) RawServiceMap {
	newServiceMap := make(RawServiceMap)
	for k, v := range serviceMap {
		newServiceMap[k] = convertServiceKeysToStrings(v)
	}
	return newServiceMap
}

func convertServiceKeysToStrings(service RawService) RawService {
	newService := make(RawService)
	for k, v := range service {
		newService[k] = utils.ConvertKeysToStrings(v)
	}
	return newService
}

var dockerConfigHints = map[string]string{
	"cpu_share":   "cpu_shares",
	"add_host":    "extra_hosts",
	"hosts":       "extra_hosts",
	"extra_host":  "extra_hosts",
	"device":      "devices",
	"link":        "links",
	"memory_swap": "memswap_limit",
	"port":        "ports",
	"privilege":   "privileged",
	"priviliged":  "privileged",
	"privilige":   "privileged",
	"volume":      "volumes",
	"workdir":     "working_dir",
}

func unsupportedConfigMessage(key string, nextErr gojsonschema.ResultError) string {
	service := serviceNameFromErrorField(nextErr.Field())

	message := fmt.Sprintf("Unsupported config option for %s service: '%s'", service, key)
	if val, ok := dockerConfigHints[key]; ok {
		message += fmt.Sprintf(" (did you mean '%s'?)", val)
	}

	return message
}

func oneOfMessage(serviceMap RawServiceMap, schema map[string]interface{}, err, nextErr gojsonschema.ResultError) string {
	switch nextErr.Type() {
	case "additional_property_not_allowed":
		property := nextErr.Details()["property"]

		return fmt.Sprintf("contains unsupported option: '%s'", property)
	case "invalid_type":
		if containsTypeError(nextErr) {
			expectedType := addArticle(nextErr.Details()["expected"].(string))

			return fmt.Sprintf("contains %s, which is an invalid type, it should be %s", getValue(serviceMap, nextErr.Context().String()), expectedType)
		}

		validTypes := parseValidTypesFromSchema(schema, err.Context().String())

		validTypesMsg := addArticle(strings.Join(validTypes, " or "))

		return fmt.Sprintf("contains an invalid type, it should be %s", validTypesMsg)
	case "unique":
		contextWithDuplicates := getValue(serviceMap, nextErr.Context().String())

		return fmt.Sprintf("contains non unique items, please remove duplicates from %s", contextWithDuplicates)
	}

	return ""
}

func invalidTypeMessage(service, key string, err gojsonschema.ResultError) string {
	expectedTypesString := err.Details()["expected"].(string)
	var expectedTypes []string

	if strings.Contains(expectedTypesString, ",") {
		expectedTypes = strings.Split(expectedTypesString[1:len(expectedTypesString)-1], ",")
	} else {
		expectedTypes = []string{expectedTypesString}
	}

	validTypesMsg := addArticle(strings.Join(expectedTypes, " or "))

	return fmt.Sprintf("Service '%s' configuration key '%s' contains an invalid type, it should be %s.", service, key, validTypesMsg)
}

func validate(serviceMap RawServiceMap) error {
	serviceMap = convertServiceMapKeysToStrings(serviceMap)

	dataLoader := gojsonschema.NewGoLoader(serviceMap)

	result, err := gojsonschema.Validate(schemaLoaderV1, dataLoader)
	if err != nil {
		return err
	}

	return generateErrorMessages(serviceMap, schemaV1, result)
}

func validateV2(serviceMap RawServiceMap) error {
	serviceMap = convertServiceMapKeysToStrings(serviceMap)

	dataLoader := gojsonschema.NewGoLoader(serviceMap)

	result, err := gojsonschema.Validate(schemaLoaderV2, dataLoader)
	if err != nil {
		return err
	}

	return generateErrorMessages(serviceMap, schemaV2, result)
}

func generateErrorMessages(serviceMap RawServiceMap, schema map[string]interface{}, result *gojsonschema.Result) error {
	var validationErrors []string

	// gojsonschema can create extraneous "additional_property_not_allowed" errors in some cases
	// If this is set, and the error is at root level, skip over that error
	skipRootAdditionalPropertyError := false

	if !result.Valid() {
		for i := 0; i < len(result.Errors()); i++ {
			err := result.Errors()[i]

			if skipRootAdditionalPropertyError && err.Type() == "additional_property_not_allowed" && err.Context().String() == "(root)" {
				skipRootAdditionalPropertyError = false
				continue
			}

			if err.Context().String() == "(root)" {
				switch err.Type() {
				case "additional_property_not_allowed":
					validationErrors = append(validationErrors, fmt.Sprintf("Invalid service name '%s' - only [a-zA-Z0-9\\._\\-] characters are allowed", err.Field()))
				default:
					validationErrors = append(validationErrors, err.Description())
				}
			} else {
				skipRootAdditionalPropertyError = true

				serviceName := serviceNameFromErrorField(err.Field())
				key := keyNameFromErrorField(err.Field())

				switch err.Type() {
				case "additional_property_not_allowed":
					validationErrors = append(validationErrors, unsupportedConfigMessage(key, result.Errors()[i+1]))
				case "number_one_of":
					validationErrors = append(validationErrors, fmt.Sprintf("Service '%s' configuration key '%s' %s", serviceName, key, oneOfMessage(serviceMap, schema, err, result.Errors()[i+1])))

					// Next error handled in oneOfMessage, skip over it
					i++
				case "invalid_type":
					validationErrors = append(validationErrors, invalidTypeMessage(serviceName, key, err))
				case "required":
					validationErrors = append(validationErrors, fmt.Sprintf("Service '%s' option '%s' is invalid, %s", serviceName, key, err.Description()))
				case "missing_dependency":
					dependency := err.Details()["dependency"].(string)
					validationErrors = append(validationErrors, fmt.Sprintf("Invalid configuration for '%s' service: dependency '%s' is not satisfied", serviceName, dependency))
				case "unique":
					contextWithDuplicates := getValue(serviceMap, err.Context().String())
					validationErrors = append(validationErrors, fmt.Sprintf("Service '%s' configuration key '%s' value %s has non-unique elements", serviceName, key, contextWithDuplicates))
				default:
					validationErrors = append(validationErrors, fmt.Sprintf("Service '%s' configuration key %s value %s", serviceName, key, err.Description()))
				}
			}
		}

		return fmt.Errorf(strings.Join(validationErrors, "\n"))
	}

	return nil
}

func validateServiceConstraints(service RawService, serviceName string) error {
	service = convertServiceKeysToStrings(service)

	var validationErrors []string

	dataLoader := gojsonschema.NewGoLoader(service)

	result, err := gojsonschema.Validate(constraintSchemaLoaderV1, dataLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			if err.Type() == "number_any_of" {
				_, containsImage := service["image"]
				_, containsBuild := service["build"]
				_, containsDockerfile := service["dockerfile"]

				if containsImage && containsBuild {
					validationErrors = append(validationErrors, fmt.Sprintf("Service '%s' has both an image and build path specified. A service can either be built to image or use an existing image, not both.", serviceName))
				} else if !containsImage && !containsBuild {
					validationErrors = append(validationErrors, fmt.Sprintf("Service '%s' has neither an image nor a build path specified. Exactly one must be provided.", serviceName))
				} else if containsImage && containsDockerfile {
					validationErrors = append(validationErrors, fmt.Sprintf("Service '%s' has both an image and alternate Dockerfile. A service can either be built to image or use an existing image, not both.", serviceName))
				}
			}
		}

		return fmt.Errorf(strings.Join(validationErrors, "\n"))
	}

	return nil
}

func validateServiceConstraintsv2(service RawService, serviceName string) error {
	service = convertServiceKeysToStrings(service)

	var validationErrors []string

	dataLoader := gojsonschema.NewGoLoader(service)

	result, err := gojsonschema.Validate(constraintSchemaLoaderV2, dataLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			if err.Type() == "required" {
				_, containsImage := service["image"]
				_, containsBuild := service["build"]

				if containsBuild || !containsImage && !containsBuild {
					validationErrors = append(validationErrors, fmt.Sprintf("Service '%s' has neither an image nor a build context specified. At least one must be provided.", serviceName))
				}
			}
		}
		return fmt.Errorf(strings.Join(validationErrors, "\n"))
	}

	return nil
}
