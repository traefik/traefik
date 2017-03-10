package config

import (
	"encoding/json"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/xeipuuv/gojsonschema"
)

var (
	schemaLoader           gojsonschema.JSONLoader
	constraintSchemaLoader gojsonschema.JSONLoader
	schema                 map[string]interface{}
)

type (
	environmentFormatChecker struct{}
	portsFormatChecker       struct{}
)

func (checker environmentFormatChecker) IsFormat(input string) bool {
	// If the value is a boolean, a warning should be given
	// However, we can't determine type since gojsonschema converts the value to a string
	// Adding a function with an interface{} parameter to gojsonschema is probably the best way to handle this
	return true
}

func (checker portsFormatChecker) IsFormat(input string) bool {
	_, _, err := nat.ParsePortSpecs([]string{input})
	return err == nil
}

func setupSchemaLoaders() error {
	if schema != nil {
		return nil
	}

	var schemaRaw interface{}
	err := json.Unmarshal([]byte(schemaV1), &schemaRaw)
	if err != nil {
		return err
	}

	schema = schemaRaw.(map[string]interface{})

	gojsonschema.FormatCheckers.Add("environment", environmentFormatChecker{})
	gojsonschema.FormatCheckers.Add("ports", portsFormatChecker{})
	gojsonschema.FormatCheckers.Add("expose", portsFormatChecker{})
	schemaLoader = gojsonschema.NewGoLoader(schemaRaw)

	definitions := schema["definitions"].(map[string]interface{})
	constraints := definitions["constraints"].(map[string]interface{})
	service := constraints["service"].(map[string]interface{})
	constraintSchemaLoader = gojsonschema.NewGoLoader(service)

	return nil
}

// gojsonschema doesn't provide a list of valid types for a property
// This parses the schema manually to find all valid types
func parseValidTypesFromSchema(schema map[string]interface{}, context string) []string {
	contextSplit := strings.Split(context, ".")
	key := contextSplit[len(contextSplit)-1]

	definitions := schema["definitions"].(map[string]interface{})
	service := definitions["service"].(map[string]interface{})
	properties := service["properties"].(map[string]interface{})
	property := properties[key].(map[string]interface{})

	var validTypes []string

	if val, ok := property["oneOf"]; ok {
		validConditions := val.([]interface{})

		for _, validCondition := range validConditions {
			condition := validCondition.(map[string]interface{})
			validTypes = append(validTypes, condition["type"].(string))
		}
	} else if val, ok := property["$ref"]; ok {
		reference := val.(string)
		if reference == "#/definitions/string_or_list" {
			return []string{"string", "array"}
		} else if reference == "#/definitions/list_of_strings" {
			return []string{"array"}
		} else if reference == "#/definitions/list_or_dict" {
			return []string{"array", "object"}
		}
	}

	return validTypes
}
