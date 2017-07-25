// Copyright 2015 xeipuuv ( https://github.com/xeipuuv )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author           xeipuuv
// author-github    https://github.com/xeipuuv
// author-mail      xeipuuv@gmail.com
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description      (Unit) Tests for schema validation.
//
// created          16-06-2013

package gojsonschema

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const displayErrorMessages = false

var rxInvoice = regexp.MustCompile("^[A-Z]{2}-[0-9]{5}")
var rxSplitErrors = regexp.MustCompile(", ?")

// Used for remote schema in ref/schema_5.json that defines "uri" and "regex" types
type alwaysTrueFormatChecker struct{}
type invoiceFormatChecker struct{}

func (a alwaysTrueFormatChecker) IsFormat(input string) bool {
	return true
}

func (a invoiceFormatChecker) IsFormat(input string) bool {
	return rxInvoice.MatchString(input)
}

func TestJsonSchemaTestSuite(t *testing.T) {

	JsonSchemaTestSuiteMap := []map[string]string{

		{"phase": "integer type matches integers", "test": "an integer is an integer", "schema": "type/schema_0.json", "data": "type/data_00.json", "valid": "true"},
		{"phase": "integer type matches integers", "test": "a float is not an integer", "schema": "type/schema_0.json", "data": "type/data_01.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "integer type matches integers", "test": "a string is not an integer", "schema": "type/schema_0.json", "data": "type/data_02.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "integer type matches integers", "test": "an object is not an integer", "schema": "type/schema_0.json", "data": "type/data_03.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "integer type matches integers", "test": "an array is not an integer", "schema": "type/schema_0.json", "data": "type/data_04.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "integer type matches integers", "test": "a boolean is not an integer", "schema": "type/schema_0.json", "data": "type/data_05.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "integer type matches integers", "test": "null is not an integer", "schema": "type/schema_0.json", "data": "type/data_06.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "number type matches numbers", "test": "an integer is a number", "schema": "type/schema_1.json", "data": "type/data_10.json", "valid": "true"},
		{"phase": "number type matches numbers", "test": "a float is a number", "schema": "type/schema_1.json", "data": "type/data_11.json", "valid": "true"},
		{"phase": "number type matches numbers", "test": "a string is not a number", "schema": "type/schema_1.json", "data": "type/data_12.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "number type matches numbers", "test": "an object is not a number", "schema": "type/schema_1.json", "data": "type/data_13.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "number type matches numbers", "test": "an array is not a number", "schema": "type/schema_1.json", "data": "type/data_14.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "number type matches numbers", "test": "a boolean is not a number", "schema": "type/schema_1.json", "data": "type/data_15.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "number type matches numbers", "test": "null is not a number", "schema": "type/schema_1.json", "data": "type/data_16.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "string type matches strings", "test": "1 is not a string", "schema": "type/schema_2.json", "data": "type/data_20.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "string type matches strings", "test": "a float is not a string", "schema": "type/schema_2.json", "data": "type/data_21.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "string type matches strings", "test": "a string is a string", "schema": "type/schema_2.json", "data": "type/data_22.json", "valid": "true"},
		{"phase": "string type matches strings", "test": "an object is not a string", "schema": "type/schema_2.json", "data": "type/data_23.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "string type matches strings", "test": "an array is not a string", "schema": "type/schema_2.json", "data": "type/data_24.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "string type matches strings", "test": "a boolean is not a string", "schema": "type/schema_2.json", "data": "type/data_25.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "string type matches strings", "test": "null is not a string", "schema": "type/schema_2.json", "data": "type/data_26.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "object type matches objects", "test": "an integer is not an object", "schema": "type/schema_3.json", "data": "type/data_30.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "object type matches objects", "test": "a float is not an object", "schema": "type/schema_3.json", "data": "type/data_31.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "object type matches objects", "test": "a string is not an object", "schema": "type/schema_3.json", "data": "type/data_32.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "object type matches objects", "test": "an object is an object", "schema": "type/schema_3.json", "data": "type/data_33.json", "valid": "true"},
		{"phase": "object type matches objects", "test": "an array is not an object", "schema": "type/schema_3.json", "data": "type/data_34.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "object type matches objects", "test": "a boolean is not an object", "schema": "type/schema_3.json", "data": "type/data_35.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "object type matches objects", "test": "null is not an object", "schema": "type/schema_3.json", "data": "type/data_36.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "array type matches arrays", "test": "an integer is not an array", "schema": "type/schema_4.json", "data": "type/data_40.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "array type matches arrays", "test": "a float is not an array", "schema": "type/schema_4.json", "data": "type/data_41.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "array type matches arrays", "test": "a string is not an array", "schema": "type/schema_4.json", "data": "type/data_42.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "array type matches arrays", "test": "an object is not an array", "schema": "type/schema_4.json", "data": "type/data_43.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "array type matches arrays", "test": "an array is not an array", "schema": "type/schema_4.json", "data": "type/data_44.json", "valid": "true"},
		{"phase": "array type matches arrays", "test": "a boolean is not an array", "schema": "type/schema_4.json", "data": "type/data_45.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "array type matches arrays", "test": "null is not an array", "schema": "type/schema_4.json", "data": "type/data_46.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "boolean type matches booleans", "test": "an integer is not a boolean", "schema": "type/schema_5.json", "data": "type/data_50.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "boolean type matches booleans", "test": "a float is not a boolean", "schema": "type/schema_5.json", "data": "type/data_51.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "boolean type matches booleans", "test": "a string is not a boolean", "schema": "type/schema_5.json", "data": "type/data_52.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "boolean type matches booleans", "test": "an object is not a boolean", "schema": "type/schema_5.json", "data": "type/data_53.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "boolean type matches booleans", "test": "an array is not a boolean", "schema": "type/schema_5.json", "data": "type/data_54.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "boolean type matches booleans", "test": "a boolean is not a boolean", "schema": "type/schema_5.json", "data": "type/data_55.json", "valid": "true", "errors": "invalid_type"},
		{"phase": "boolean type matches booleans", "test": "null is not a boolean", "schema": "type/schema_5.json", "data": "type/data_56.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "null type matches only the null object", "test": "an integer is not null", "schema": "type/schema_6.json", "data": "type/data_60.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "null type matches only the null object", "test": "a float is not null", "schema": "type/schema_6.json", "data": "type/data_61.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "null type matches only the null object", "test": "a string is not null", "schema": "type/schema_6.json", "data": "type/data_62.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "null type matches only the null object", "test": "an object is not null", "schema": "type/schema_6.json", "data": "type/data_63.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "null type matches only the null object", "test": "an array is not null", "schema": "type/schema_6.json", "data": "type/data_64.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "null type matches only the null object", "test": "a boolean is not null", "schema": "type/schema_6.json", "data": "type/data_65.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "null type matches only the null object", "test": "null is null", "schema": "type/schema_6.json", "data": "type/data_66.json", "valid": "true"},
		{"phase": "multiple types can be specified in an array", "test": "an integer is valid", "schema": "type/schema_7.json", "data": "type/data_70.json", "valid": "true"},
		{"phase": "multiple types can be specified in an array", "test": "a string is valid", "schema": "type/schema_7.json", "data": "type/data_71.json", "valid": "true"},
		{"phase": "multiple types can be specified in an array", "test": "a float is invalid", "schema": "type/schema_7.json", "data": "type/data_72.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "multiple types can be specified in an array", "test": "an object is invalid", "schema": "type/schema_7.json", "data": "type/data_73.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "multiple types can be specified in an array", "test": "an array is invalid", "schema": "type/schema_7.json", "data": "type/data_74.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "multiple types can be specified in an array", "test": "a boolean is invalid", "schema": "type/schema_7.json", "data": "type/data_75.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "multiple types can be specified in an array", "test": "null is invalid", "schema": "type/schema_7.json", "data": "type/data_76.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "required validation", "test": "present required property is valid", "schema": "required/schema_0.json", "data": "required/data_00.json", "valid": "true"},
		{"phase": "required validation", "test": "non-present required property is invalid", "schema": "required/schema_0.json", "data": "required/data_01.json", "valid": "false", "errors": "required"},
		{"phase": "required default validation", "test": "not required by default", "schema": "required/schema_1.json", "data": "required/data_10.json", "valid": "true"},
		{"phase": "uniqueItems validation", "test": "unique array of integers is valid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_00.json", "valid": "true"},
		{"phase": "uniqueItems validation", "test": "non-unique array of integers is invalid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_01.json", "valid": "false", "errors": "unique"},
		{"phase": "uniqueItems validation", "test": "numbers are unique if mathematically unequal", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_02.json", "valid": "false", "errors": "unique, unique"},
		{"phase": "uniqueItems validation", "test": "unique array of objects is valid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_03.json", "valid": "true"},
		{"phase": "uniqueItems validation", "test": "non-unique array of objects is invalid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_04.json", "valid": "false", "errors": "unique"},
		{"phase": "uniqueItems validation", "test": "unique array of nested objects is valid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_05.json", "valid": "true"},
		{"phase": "uniqueItems validation", "test": "non-unique array of nested objects is invalid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_06.json", "valid": "false", "errors": "unique"},
		{"phase": "uniqueItems validation", "test": "unique array of arrays is valid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_07.json", "valid": "true"},
		{"phase": "uniqueItems validation", "test": "non-unique array of arrays is invalid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_08.json", "valid": "false", "errors": "unique"},
		{"phase": "uniqueItems validation", "test": "1 and true are unique", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_09.json", "valid": "true"},
		{"phase": "uniqueItems validation", "test": "0 and false are unique", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_010.json", "valid": "true"},
		{"phase": "uniqueItems validation", "test": "unique heterogeneous types are valid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_011.json", "valid": "true"},
		{"phase": "uniqueItems validation", "test": "non-unique heterogeneous types are invalid", "schema": "uniqueItems/schema_0.json", "data": "uniqueItems/data_012.json", "valid": "false", "errors": "unique"},
		{"phase": "pattern validation", "test": "a matching pattern is valid", "schema": "pattern/schema_0.json", "data": "pattern/data_00.json", "valid": "true"},
		{"phase": "pattern validation", "test": "a non-matching pattern is invalid", "schema": "pattern/schema_0.json", "data": "pattern/data_01.json", "valid": "false", "errors": "pattern"},
		{"phase": "pattern validation", "test": "ignores non-strings", "schema": "pattern/schema_0.json", "data": "pattern/data_02.json", "valid": "true"},
		{"phase": "simple enum validation", "test": "one of the enum is valid", "schema": "enum/schema_0.json", "data": "enum/data_00.json", "valid": "true"},
		{"phase": "simple enum validation", "test": "something else is invalid", "schema": "enum/schema_0.json", "data": "enum/data_01.json", "valid": "false", "errors": "enum"},
		{"phase": "heterogeneous enum validation", "test": "one of the enum is valid", "schema": "enum/schema_1.json", "data": "enum/data_10.json", "valid": "true"},
		{"phase": "heterogeneous enum validation", "test": "something else is invalid", "schema": "enum/schema_1.json", "data": "enum/data_11.json", "valid": "false", "errors": "enum"},
		{"phase": "heterogeneous enum validation", "test": "objects are deep compared", "schema": "enum/schema_1.json", "data": "enum/data_12.json", "valid": "false", "errors": "enum"},
		{"phase": "minLength validation", "test": "longer is valid", "schema": "minLength/schema_0.json", "data": "minLength/data_00.json", "valid": "true"},
		{"phase": "minLength validation", "test": "exact length is valid", "schema": "minLength/schema_0.json", "data": "minLength/data_01.json", "valid": "true"},
		{"phase": "minLength validation", "test": "too short is invalid", "schema": "minLength/schema_0.json", "data": "minLength/data_02.json", "valid": "false", "errors": "string_gte"},
		{"phase": "minLength validation", "test": "ignores non-strings", "schema": "minLength/schema_0.json", "data": "minLength/data_03.json", "valid": "true"},
		{"phase": "minLength validation", "test": "counts utf8 length correctly", "schema": "minLength/schema_0.json", "data": "minLength/data_04.json", "valid": "false", "errors": "string_gte"},
		{"phase": "maxLength validation", "test": "shorter is valid", "schema": "maxLength/schema_0.json", "data": "maxLength/data_00.json", "valid": "true"},
		{"phase": "maxLength validation", "test": "exact length is valid", "schema": "maxLength/schema_0.json", "data": "maxLength/data_01.json", "valid": "true"},
		{"phase": "maxLength validation", "test": "too long is invalid", "schema": "maxLength/schema_0.json", "data": "maxLength/data_02.json", "valid": "false", "errors": "string_lte"},
		{"phase": "maxLength validation", "test": "ignores non-strings", "schema": "maxLength/schema_0.json", "data": "maxLength/data_03.json", "valid": "true"},
		{"phase": "maxLength validation", "test": "counts utf8 length correctly", "schema": "maxLength/schema_0.json", "data": "maxLength/data_04.json", "valid": "true"},
		{"phase": "minimum validation", "test": "above the minimum is valid", "schema": "minimum/schema_0.json", "data": "minimum/data_00.json", "valid": "true"},
		{"phase": "minimum validation", "test": "below the minimum is invalid", "schema": "minimum/schema_0.json", "data": "minimum/data_01.json", "valid": "false", "errors": "number_gte"},
		{"phase": "minimum validation", "test": "ignores non-numbers", "schema": "minimum/schema_0.json", "data": "minimum/data_02.json", "valid": "true"},
		{"phase": "exclusiveMinimum validation", "test": "above the minimum is still valid", "schema": "minimum/schema_1.json", "data": "minimum/data_10.json", "valid": "true"},
		{"phase": "exclusiveMinimum validation", "test": "boundary point is invalid", "schema": "minimum/schema_1.json", "data": "minimum/data_11.json", "valid": "false", "errors": "number_gt"},
		{"phase": "maximum validation", "test": "below the maximum is valid", "schema": "maximum/schema_0.json", "data": "maximum/data_00.json", "valid": "true"},
		{"phase": "maximum validation", "test": "above the maximum is invalid", "schema": "maximum/schema_0.json", "data": "maximum/data_01.json", "valid": "false", "errors": "number_lte"},
		{"phase": "maximum validation", "test": "ignores non-numbers", "schema": "maximum/schema_0.json", "data": "maximum/data_02.json", "valid": "true"},
		{"phase": "exclusiveMaximum validation", "test": "below the maximum is still valid", "schema": "maximum/schema_1.json", "data": "maximum/data_10.json", "valid": "true"},
		{"phase": "exclusiveMaximum validation", "test": "boundary point is invalid", "schema": "maximum/schema_1.json", "data": "maximum/data_11.json", "valid": "false", "errors": "number_lt"},
		{"phase": "allOf", "test": "allOf", "schema": "allOf/schema_0.json", "data": "allOf/data_00.json", "valid": "true"},
		{"phase": "allOf", "test": "mismatch second", "schema": "allOf/schema_0.json", "data": "allOf/data_01.json", "valid": "false", "errors": "number_all_of, required"},
		{"phase": "allOf", "test": "mismatch first", "schema": "allOf/schema_0.json", "data": "allOf/data_02.json", "valid": "false", "errors": "number_all_of, required"},
		{"phase": "allOf", "test": "wrong type", "schema": "allOf/schema_0.json", "data": "allOf/data_03.json", "valid": "false", "errors": "number_all_of, invalid_type"},
		{"phase": "allOf with base schema", "test": "valid", "schema": "allOf/schema_1.json", "data": "allOf/data_10.json", "valid": "true"},
		{"phase": "allOf with base schema", "test": "mismatch base schema", "schema": "allOf/schema_1.json", "data": "allOf/data_11.json", "valid": "false", "errors": "required"},
		{"phase": "allOf with base schema", "test": "mismatch first allOf", "schema": "allOf/schema_1.json", "data": "allOf/data_12.json", "valid": "false", "errors": "number_all_of, required"},
		{"phase": "allOf with base schema", "test": "mismatch second allOf", "schema": "allOf/schema_1.json", "data": "allOf/data_13.json", "valid": "false", "errors": "number_all_of, required"},
		{"phase": "allOf with base schema", "test": "mismatch both", "schema": "allOf/schema_1.json", "data": "allOf/data_14.json", "valid": "false", "errors": "number_all_of, required, required"},
		{"phase": "allOf simple types", "test": "valid", "schema": "allOf/schema_2.json", "data": "allOf/data_20.json", "valid": "true"},
		{"phase": "allOf simple types", "test": "mismatch one", "schema": "allOf/schema_2.json", "data": "allOf/data_21.json", "valid": "false", "errors": "number_all_of, number_lte"},
		{"phase": "oneOf", "test": "first oneOf valid", "schema": "oneOf/schema_0.json", "data": "oneOf/data_00.json", "valid": "true"},
		{"phase": "oneOf", "test": "second oneOf valid", "schema": "oneOf/schema_0.json", "data": "oneOf/data_01.json", "valid": "true"},
		{"phase": "oneOf", "test": "both oneOf valid", "schema": "oneOf/schema_0.json", "data": "oneOf/data_02.json", "valid": "false", "errors": "number_one_of"},
		{"phase": "oneOf", "test": "neither oneOf valid", "schema": "oneOf/schema_0.json", "data": "oneOf/data_03.json", "valid": "false"},
		{"phase": "oneOf with base schema", "test": "mismatch base schema", "schema": "oneOf/schema_1.json", "data": "oneOf/data_10.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "oneOf with base schema", "test": "one oneOf valid", "schema": "oneOf/schema_1.json", "data": "oneOf/data_11.json", "valid": "true"},
		{"phase": "oneOf with base schema", "test": "both oneOf valid", "schema": "oneOf/schema_1.json", "data": "oneOf/data_12.json", "valid": "false", "errors": "number_one_of"},
		{"phase": "anyOf", "test": "first anyOf valid", "schema": "anyOf/schema_0.json", "data": "anyOf/data_00.json", "valid": "true"},
		{"phase": "anyOf", "test": "second anyOf valid", "schema": "anyOf/schema_0.json", "data": "anyOf/data_01.json", "valid": "true"},
		{"phase": "anyOf", "test": "both anyOf valid", "schema": "anyOf/schema_0.json", "data": "anyOf/data_02.json", "valid": "true"},
		{"phase": "anyOf", "test": "neither anyOf valid", "schema": "anyOf/schema_0.json", "data": "anyOf/data_03.json", "valid": "false", "errors": "number_any_of, number_gte"},
		{"phase": "anyOf with base schema", "test": "mismatch base schema", "schema": "anyOf/schema_1.json", "data": "anyOf/data_10.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "anyOf with base schema", "test": "one anyOf valid", "schema": "anyOf/schema_1.json", "data": "anyOf/data_11.json", "valid": "true"},
		{"phase": "anyOf with base schema", "test": "both anyOf invalid", "schema": "anyOf/schema_1.json", "data": "anyOf/data_12.json", "valid": "false", "errors": "number_any_of, string_lte"},
		{"phase": "not", "test": "allowed", "schema": "not/schema_0.json", "data": "not/data_00.json", "valid": "true"},
		{"phase": "not", "test": "disallowed", "schema": "not/schema_0.json", "data": "not/data_01.json", "valid": "false", "errors": "number_not"},
		{"phase": "not multiple types", "test": "valid", "schema": "not/schema_1.json", "data": "not/data_10.json", "valid": "true"},
		{"phase": "not multiple types", "test": "mismatch", "schema": "not/schema_1.json", "data": "not/data_11.json", "valid": "false", "errors": "number_not"},
		{"phase": "not multiple types", "test": "other mismatch", "schema": "not/schema_1.json", "data": "not/data_12.json", "valid": "false", "errors": "number_not"},
		{"phase": "not more complex schema", "test": "match", "schema": "not/schema_2.json", "data": "not/data_20.json", "valid": "true"},
		{"phase": "not more complex schema", "test": "other match", "schema": "not/schema_2.json", "data": "not/data_21.json", "valid": "true"},
		{"phase": "not more complex schema", "test": "mismatch", "schema": "not/schema_2.json", "data": "not/data_22.json", "valid": "false", "errors": "number_not"},
		{"phase": "minProperties validation", "test": "longer is valid", "schema": "minProperties/schema_0.json", "data": "minProperties/data_00.json", "valid": "true"},
		{"phase": "minProperties validation", "test": "exact length is valid", "schema": "minProperties/schema_0.json", "data": "minProperties/data_01.json", "valid": "true"},
		{"phase": "minProperties validation", "test": "too short is invalid", "schema": "minProperties/schema_0.json", "data": "minProperties/data_02.json", "valid": "false", "errors": "array_min_properties"},
		{"phase": "minProperties validation", "test": "ignores non-objects", "schema": "minProperties/schema_0.json", "data": "minProperties/data_03.json", "valid": "true"},
		{"phase": "maxProperties validation", "test": "shorter is valid", "schema": "maxProperties/schema_0.json", "data": "maxProperties/data_00.json", "valid": "true"},
		{"phase": "maxProperties validation", "test": "exact length is valid", "schema": "maxProperties/schema_0.json", "data": "maxProperties/data_01.json", "valid": "true"},
		{"phase": "maxProperties validation", "test": "too long is invalid", "schema": "maxProperties/schema_0.json", "data": "maxProperties/data_02.json", "valid": "false", "errors": "array_max_properties"},
		{"phase": "maxProperties validation", "test": "ignores non-objects", "schema": "maxProperties/schema_0.json", "data": "maxProperties/data_03.json", "valid": "true"},
		{"phase": "by int", "test": "int by int", "schema": "multipleOf/schema_0.json", "data": "multipleOf/data_00.json", "valid": "true"},
		{"phase": "by int", "test": "int by int fail", "schema": "multipleOf/schema_0.json", "data": "multipleOf/data_01.json", "valid": "false", "errors": "multiple_of"},
		{"phase": "by int", "test": "ignores non-numbers", "schema": "multipleOf/schema_0.json", "data": "multipleOf/data_02.json", "valid": "true"},
		{"phase": "by number", "test": "zero is multiple of anything", "schema": "multipleOf/schema_1.json", "data": "multipleOf/data_10.json", "valid": "true"},
		{"phase": "by number", "test": "4.5 is multiple of 1.5", "schema": "multipleOf/schema_1.json", "data": "multipleOf/data_11.json", "valid": "true"},
		{"phase": "by number", "test": "35 is not multiple of 1.5", "schema": "multipleOf/schema_1.json", "data": "multipleOf/data_12.json", "valid": "false", "errors": "multiple_of"},
		{"phase": "by small number", "test": "0.0075 is multiple of 0.0001", "schema": "multipleOf/schema_2.json", "data": "multipleOf/data_20.json", "valid": "true"},
		{"phase": "by small number", "test": "0.00751 is not multiple of 0.0001", "schema": "multipleOf/schema_2.json", "data": "multipleOf/data_21.json", "valid": "false", "errors": "multiple_of"},
		{"phase": "minItems validation", "test": "longer is valid", "schema": "minItems/schema_0.json", "data": "minItems/data_00.json", "valid": "true"},
		{"phase": "minItems validation", "test": "exact length is valid", "schema": "minItems/schema_0.json", "data": "minItems/data_01.json", "valid": "true"},
		{"phase": "minItems validation", "test": "too short is invalid", "schema": "minItems/schema_0.json", "data": "minItems/data_02.json", "valid": "false", "errors": "array_min_items"},
		{"phase": "minItems validation", "test": "ignores non-arrays", "schema": "minItems/schema_0.json", "data": "minItems/data_03.json", "valid": "true"},
		{"phase": "maxItems validation", "test": "shorter is valid", "schema": "maxItems/schema_0.json", "data": "maxItems/data_00.json", "valid": "true"},
		{"phase": "maxItems validation", "test": "exact length is valid", "schema": "maxItems/schema_0.json", "data": "maxItems/data_01.json", "valid": "true"},
		{"phase": "maxItems validation", "test": "too long is invalid", "schema": "maxItems/schema_0.json", "data": "maxItems/data_02.json", "valid": "false", "errors": "array_max_items"},
		{"phase": "maxItems validation", "test": "ignores non-arrays", "schema": "maxItems/schema_0.json", "data": "maxItems/data_03.json", "valid": "true"},
		{"phase": "object properties validation", "test": "both properties present and valid is valid", "schema": "properties/schema_0.json", "data": "properties/data_00.json", "valid": "true"},
		{"phase": "object properties validation", "test": "one property invalid is invalid", "schema": "properties/schema_0.json", "data": "properties/data_01.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "object properties validation", "test": "both properties invalid is invalid", "schema": "properties/schema_0.json", "data": "properties/data_02.json", "valid": "false", "errors": "invalid_type, invalid_type"},
		{"phase": "object properties validation", "test": "doesn't invalidate other properties", "schema": "properties/schema_0.json", "data": "properties/data_03.json", "valid": "true"},
		{"phase": "object properties validation", "test": "ignores non-objects", "schema": "properties/schema_0.json", "data": "properties/data_04.json", "valid": "true"},
		{"phase": "properties, patternProperties, additionalProperties interaction", "test": "property validates property", "schema": "properties/schema_1.json", "data": "properties/data_10.json", "valid": "true"},
		{"phase": "properties, patternProperties, additionalProperties interaction", "test": "property invalidates property", "schema": "properties/schema_1.json", "data": "properties/data_11.json", "valid": "false", "errors": "array_max_items"},
		{"phase": "properties, patternProperties, additionalProperties interaction", "test": "patternProperty invalidates property", "schema": "properties/schema_1.json", "data": "properties/data_12.json", "valid": "false", "errors": "array_min_items, invalid_type"},
		{"phase": "properties, patternProperties, additionalProperties interaction", "test": "patternProperty validates nonproperty", "schema": "properties/schema_1.json", "data": "properties/data_13.json", "valid": "true"},
		{"phase": "properties, patternProperties, additionalProperties interaction", "test": "patternProperty invalidates nonproperty", "schema": "properties/schema_1.json", "data": "properties/data_14.json", "valid": "false", "errors": "array_min_items, invalid_type"},
		{"phase": "properties, patternProperties, additionalProperties interaction", "test": "additionalProperty ignores property", "schema": "properties/schema_1.json", "data": "properties/data_15.json", "valid": "true"},
		{"phase": "properties, patternProperties, additionalProperties interaction", "test": "additionalProperty validates others", "schema": "properties/schema_1.json", "data": "properties/data_16.json", "valid": "true"},
		{"phase": "properties, patternProperties, additionalProperties interaction", "test": "additionalProperty invalidates others", "schema": "properties/schema_1.json", "data": "properties/data_17.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "root pointer ref", "test": "match", "schema": "ref/schema_0.json", "data": "ref/data_00.json", "valid": "true"},
		{"phase": "root pointer ref", "test": "recursive match", "schema": "ref/schema_0.json", "data": "ref/data_01.json", "valid": "true"},
		{"phase": "root pointer ref", "test": "mismatch", "schema": "ref/schema_0.json", "data": "ref/data_02.json", "valid": "false", "errors": "additional_property_not_allowed"},
		{"phase": "root pointer ref", "test": "recursive mismatch", "schema": "ref/schema_0.json", "data": "ref/data_03.json", "valid": "false", "errors": "additional_property_not_allowed"},
		{"phase": "relative pointer ref to object", "test": "match", "schema": "ref/schema_1.json", "data": "ref/data_10.json", "valid": "true"},
		{"phase": "relative pointer ref to object", "test": "mismatch", "schema": "ref/schema_1.json", "data": "ref/data_11.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "relative pointer ref to array", "test": "match array", "schema": "ref/schema_2.json", "data": "ref/data_20.json", "valid": "true"},
		{"phase": "relative pointer ref to array", "test": "mismatch array", "schema": "ref/schema_2.json", "data": "ref/data_21.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "escaped pointer ref", "test": "slash", "schema": "ref/schema_3.json", "data": "ref/data_30.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "escaped pointer ref", "test": "tilda", "schema": "ref/schema_3.json", "data": "ref/data_31.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "escaped pointer ref", "test": "percent", "schema": "ref/schema_3.json", "data": "ref/data_32.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "nested refs", "test": "nested ref valid", "schema": "ref/schema_4.json", "data": "ref/data_40.json", "valid": "true"},
		{"phase": "nested refs", "test": "nested ref invalid", "schema": "ref/schema_4.json", "data": "ref/data_41.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "remote ref, containing refs itself", "test": "remote ref valid", "schema": "ref/schema_5.json", "data": "ref/data_50.json", "valid": "true"},
		{"phase": "remote ref, containing refs itself", "test": "remote ref invalid", "schema": "ref/schema_5.json", "data": "ref/data_51.json", "valid": "false", "errors": "number_all_of, number_gte"},
		{"phase": "a schema given for items", "test": "valid items", "schema": "items/schema_0.json", "data": "items/data_00.json", "valid": "true"},
		{"phase": "a schema given for items", "test": "wrong type of items", "schema": "items/schema_0.json", "data": "items/data_01.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "a schema given for items", "test": "ignores non-arrays", "schema": "items/schema_0.json", "data": "items/data_02.json", "valid": "true"},
		{"phase": "an array of schemas for items", "test": "correct types", "schema": "items/schema_1.json", "data": "items/data_10.json", "valid": "true"},
		{"phase": "an array of schemas for items", "test": "wrong types", "schema": "items/schema_1.json", "data": "items/data_11.json", "valid": "false", "errors": "invalid_type, invalid_type"},
		{"phase": "valid definition", "test": "valid definition schema", "schema": "definitions/schema_0.json", "data": "definitions/data_00.json", "valid": "true"},
		{"phase": "invalid definition", "test": "invalid definition schema", "schema": "definitions/schema_1.json", "data": "definitions/data_10.json", "valid": "false", "errors": "number_any_of, enum"},
		{"phase": "additionalItems as schema", "test": "additional items match schema", "schema": "additionalItems/schema_0.json", "data": "additionalItems/data_00.json", "valid": "true"},
		{"phase": "additionalItems as schema", "test": "additional items do not match schema", "schema": "additionalItems/schema_0.json", "data": "additionalItems/data_01.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "items is schema, no additionalItems", "test": "all items match schema", "schema": "additionalItems/schema_1.json", "data": "additionalItems/data_10.json", "valid": "true"},
		{"phase": "array of items with no additionalItems", "test": "no additional items present", "schema": "additionalItems/schema_2.json", "data": "additionalItems/data_20.json", "valid": "true"},
		{"phase": "array of items with no additionalItems", "test": "additional items are not permitted", "schema": "additionalItems/schema_2.json", "data": "additionalItems/data_21.json", "valid": "false", "errors": "array_no_additional_items"},
		{"phase": "additionalItems as false without items", "test": "items defaults to empty schema so everything is valid", "schema": "additionalItems/schema_3.json", "data": "additionalItems/data_30.json", "valid": "true"},
		{"phase": "additionalItems as false without items", "test": "ignores non-arrays", "schema": "additionalItems/schema_3.json", "data": "additionalItems/data_31.json", "valid": "true"},
		{"phase": "additionalItems are allowed by default", "test": "only the first item is validated", "schema": "additionalItems/schema_4.json", "data": "additionalItems/data_40.json", "valid": "true"},
		{"phase": "additionalProperties being false does not allow other properties", "test": "no additional properties is valid", "schema": "additionalProperties/schema_0.json", "data": "additionalProperties/data_00.json", "valid": "true"},
		{"phase": "additionalProperties being false does not allow other properties", "test": "an additional property is invalid", "schema": "additionalProperties/schema_0.json", "data": "additionalProperties/data_01.json", "valid": "false", "errors": "additional_property_not_allowed"},
		{"phase": "additionalProperties being false does not allow other properties", "test": "ignores non-objects", "schema": "additionalProperties/schema_0.json", "data": "additionalProperties/data_02.json", "valid": "true"},
		{"phase": "additionalProperties being false does not allow other properties", "test": "patternProperties are not additional properties", "schema": "additionalProperties/schema_0.json", "data": "additionalProperties/data_03.json", "valid": "true"},
		{"phase": "additionalProperties allows a schema which should validate", "test": "no additional properties is valid", "schema": "additionalProperties/schema_1.json", "data": "additionalProperties/data_10.json", "valid": "true"},
		{"phase": "additionalProperties allows a schema which should validate", "test": "an additional valid property is valid", "schema": "additionalProperties/schema_1.json", "data": "additionalProperties/data_11.json", "valid": "true"},
		{"phase": "additionalProperties allows a schema which should validate", "test": "an additional invalid property is invalid", "schema": "additionalProperties/schema_1.json", "data": "additionalProperties/data_12.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "additionalProperties are allowed by default", "test": "additional properties are allowed", "schema": "additionalProperties/schema_2.json", "data": "additionalProperties/data_20.json", "valid": "true"},
		{"phase": "dependencies", "test": "neither", "schema": "dependencies/schema_0.json", "data": "dependencies/data_00.json", "valid": "true"},
		{"phase": "dependencies", "test": "nondependant", "schema": "dependencies/schema_0.json", "data": "dependencies/data_01.json", "valid": "true"},
		{"phase": "dependencies", "test": "with dependency", "schema": "dependencies/schema_0.json", "data": "dependencies/data_02.json", "valid": "true"},
		{"phase": "dependencies", "test": "missing dependency", "schema": "dependencies/schema_0.json", "data": "dependencies/data_03.json", "valid": "false", "errors": "missing_dependency"},
		{"phase": "dependencies", "test": "ignores non-objects", "schema": "dependencies/schema_0.json", "data": "dependencies/data_04.json", "valid": "true"},
		{"phase": "multiple dependencies", "test": "neither", "schema": "dependencies/schema_1.json", "data": "dependencies/data_10.json", "valid": "true"},
		{"phase": "multiple dependencies", "test": "nondependants", "schema": "dependencies/schema_1.json", "data": "dependencies/data_11.json", "valid": "true"},
		{"phase": "multiple dependencies", "test": "with dependencies", "schema": "dependencies/schema_1.json", "data": "dependencies/data_12.json", "valid": "true"},
		{"phase": "multiple dependencies", "test": "missing dependency", "schema": "dependencies/schema_1.json", "data": "dependencies/data_13.json", "valid": "false", "errors": "missing_dependency"},
		{"phase": "multiple dependencies", "test": "missing other dependency", "schema": "dependencies/schema_1.json", "data": "dependencies/data_14.json", "valid": "false", "errors": "missing_dependency"},
		{"phase": "multiple dependencies", "test": "missing both dependencies", "schema": "dependencies/schema_1.json", "data": "dependencies/data_15.json", "valid": "false", "errors": "missing_dependency, missing_dependency"},
		{"phase": "multiple dependencies subschema", "test": "valid", "schema": "dependencies/schema_2.json", "data": "dependencies/data_20.json", "valid": "true"},
		{"phase": "multiple dependencies subschema", "test": "no dependency", "schema": "dependencies/schema_2.json", "data": "dependencies/data_21.json", "valid": "true"},
		{"phase": "multiple dependencies subschema", "test": "wrong type", "schema": "dependencies/schema_2.json", "data": "dependencies/data_22.json", "valid": "false"},
		{"phase": "multiple dependencies subschema", "test": "wrong type other", "schema": "dependencies/schema_2.json", "data": "dependencies/data_23.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "multiple dependencies subschema", "test": "wrong type both", "schema": "dependencies/schema_2.json", "data": "dependencies/data_24.json", "valid": "false", "errors": "invalid_type, invalid_type"},
		{"phase": "patternProperties validates properties matching a regex", "test": "a single valid match is valid", "schema": "patternProperties/schema_0.json", "data": "patternProperties/data_00.json", "valid": "true"},
		{"phase": "patternProperties validates properties matching a regex", "test": "multiple valid matches is valid", "schema": "patternProperties/schema_0.json", "data": "patternProperties/data_01.json", "valid": "true"},
		{"phase": "patternProperties validates properties matching a regex", "test": "a single invalid match is invalid", "schema": "patternProperties/schema_0.json", "data": "patternProperties/data_02.json", "valid": "false", "errors": "invalid_property_pattern, invalid_type"},
		{"phase": "patternProperties validates properties matching a regex", "test": "multiple invalid matches is invalid", "schema": "patternProperties/schema_0.json", "data": "patternProperties/data_03.json", "valid": "false", "errors": "invalid_property_pattern, invalid_property_pattern, invalid_type, invalid_type"},
		{"phase": "patternProperties validates properties matching a regex", "test": "ignores non-objects", "schema": "patternProperties/schema_0.json", "data": "patternProperties/data_04.json", "valid": "true"},
		{"phase": "patternProperties validates properties matching a regex", "test": "with additionalProperties combination", "schema": "patternProperties/schema_3.json", "data": "patternProperties/data_24.json", "valid": "false", "errors": "additional_property_not_allowed"},
		{"phase": "patternProperties validates properties matching a regex", "test": "with additionalProperties combination", "schema": "patternProperties/schema_3.json", "data": "patternProperties/data_25.json", "valid": "false", "errors": "additional_property_not_allowed"},
		{"phase": "patternProperties validates properties matching a regex", "test": "with additionalProperties combination", "schema": "patternProperties/schema_4.json", "data": "patternProperties/data_26.json", "valid": "false", "errors": "additional_property_not_allowed"},
		{"phase": "multiple simultaneous patternProperties are validated", "test": "a single valid match is valid", "schema": "patternProperties/schema_1.json", "data": "patternProperties/data_10.json", "valid": "true"},
		{"phase": "multiple simultaneous patternProperties are validated", "test": "a simultaneous match is valid", "schema": "patternProperties/schema_1.json", "data": "patternProperties/data_11.json", "valid": "true"},
		{"phase": "multiple simultaneous patternProperties are validated", "test": "multiple matches is valid", "schema": "patternProperties/schema_1.json", "data": "patternProperties/data_12.json", "valid": "true"},
		{"phase": "multiple simultaneous patternProperties are validated", "test": "an invalid due to one is invalid", "schema": "patternProperties/schema_1.json", "data": "patternProperties/data_13.json", "valid": "false", "errors": "invalid_property_pattern, invalid_type"},
		{"phase": "multiple simultaneous patternProperties are validated", "test": "an invalid due to the other is invalid", "schema": "patternProperties/schema_1.json", "data": "patternProperties/data_14.json", "valid": "false", "errors": "number_lte"},
		{"phase": "multiple simultaneous patternProperties are validated", "test": "an invalid due to both is invalid", "schema": "patternProperties/schema_1.json", "data": "patternProperties/data_15.json", "valid": "false", "errors": "invalid_type, number_lte"},
		{"phase": "regexes are not anchored by default and are case sensitive", "test": "non recognized members are ignored", "schema": "patternProperties/schema_2.json", "data": "patternProperties/data_20.json", "valid": "true"},
		{"phase": "regexes are not anchored by default and are case sensitive", "test": "recognized members are accounted for", "schema": "patternProperties/schema_2.json", "data": "patternProperties/data_21.json", "valid": "false", "errors": "invalid_property_pattern, invalid_type"},
		{"phase": "regexes are not anchored by default and are case sensitive", "test": "regexes are case sensitive", "schema": "patternProperties/schema_2.json", "data": "patternProperties/data_22.json", "valid": "true"},
		{"phase": "regexes are not anchored by default and are case sensitive", "test": "regexes are case sensitive, 2", "schema": "patternProperties/schema_2.json", "data": "patternProperties/data_23.json", "valid": "false", "errors": "invalid_property_pattern, invalid_type"},
		{"phase": "remote ref", "test": "remote ref valid", "schema": "refRemote/schema_0.json", "data": "refRemote/data_00.json", "valid": "true"},
		{"phase": "remote ref", "test": "remote ref invalid", "schema": "refRemote/schema_0.json", "data": "refRemote/data_01.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "fragment within remote ref", "test": "remote fragment valid", "schema": "refRemote/schema_1.json", "data": "refRemote/data_10.json", "valid": "true"},
		{"phase": "fragment within remote ref", "test": "remote fragment invalid", "schema": "refRemote/schema_1.json", "data": "refRemote/data_11.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "ref within remote ref", "test": "ref within ref valid", "schema": "refRemote/schema_2.json", "data": "refRemote/data_20.json", "valid": "true"},
		{"phase": "ref within remote ref", "test": "ref within ref invalid", "schema": "refRemote/schema_2.json", "data": "refRemote/data_21.json", "valid": "false", "errors": "invalid_type"},
		{"phase": "format validation", "test": "email format is invalid", "schema": "format/schema_0.json", "data": "format/data_00.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "email format is invalid", "schema": "format/schema_0.json", "data": "format/data_01.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "email format valid", "schema": "format/schema_0.json", "data": "format/data_02.json", "valid": "true"},
		{"phase": "format validation", "test": "invoice format valid", "schema": "format/schema_1.json", "data": "format/data_03.json", "valid": "true"},
		{"phase": "format validation", "test": "invoice format is invalid", "schema": "format/schema_1.json", "data": "format/data_04.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "date-time format is valid", "schema": "format/schema_2.json", "data": "format/data_05.json", "valid": "true"},
		{"phase": "format validation", "test": "date-time format is valid", "schema": "format/schema_2.json", "data": "format/data_06.json", "valid": "true"},
		{"phase": "format validation", "test": "date-time format is invalid", "schema": "format/schema_2.json", "data": "format/data_07.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "date-time format is valid", "schema": "format/schema_2.json", "data": "format/data_08.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "date-time format is valid", "schema": "format/schema_2.json", "data": "format/data_09.json", "valid": "true"},
		{"phase": "format validation", "test": "date-time format is valid", "schema": "format/schema_2.json", "data": "format/data_10.json", "valid": "true"},
		{"phase": "format validation", "test": "date-time format is valid", "schema": "format/schema_2.json", "data": "format/data_11.json", "valid": "true"},
		{"phase": "format validation", "test": "date-time format is valid", "schema": "format/schema_2.json", "data": "format/data_12.json", "valid": "true"},
		{"phase": "format validation", "test": "hostname format is valid", "schema": "format/schema_3.json", "data": "format/data_13.json", "valid": "true"},
		{"phase": "format validation", "test": "hostname format is valid", "schema": "format/schema_3.json", "data": "format/data_14.json", "valid": "true"},
		{"phase": "format validation", "test": "hostname format is valid", "schema": "format/schema_3.json", "data": "format/data_15.json", "valid": "true"},
		{"phase": "format validation", "test": "hostname format is invalid", "schema": "format/schema_3.json", "data": "format/data_16.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "hostname format is invalid", "schema": "format/schema_3.json", "data": "format/data_17.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "ipv4 format is valid", "schema": "format/schema_4.json", "data": "format/data_18.json", "valid": "true"},
		{"phase": "format validation", "test": "ipv4 format is invalid", "schema": "format/schema_4.json", "data": "format/data_19.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "ipv6 format is valid", "schema": "format/schema_5.json", "data": "format/data_20.json", "valid": "true"},
		{"phase": "format validation", "test": "ipv6 format is valid", "schema": "format/schema_5.json", "data": "format/data_21.json", "valid": "true"},
		{"phase": "format validation", "test": "ipv6 format is invalid", "schema": "format/schema_5.json", "data": "format/data_22.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "ipv6 format is invalid", "schema": "format/schema_5.json", "data": "format/data_23.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "uri format is valid", "schema": "format/schema_6.json", "data": "format/data_24.json", "valid": "true"},
		{"phase": "format validation", "test": "uri format is valid", "schema": "format/schema_6.json", "data": "format/data_25.json", "valid": "true"},
		{"phase": "format validation", "test": "uri format is valid", "schema": "format/schema_6.json", "data": "format/data_26.json", "valid": "true"},
		{"phase": "format validation", "test": "uri format is valid", "schema": "format/schema_6.json", "data": "format/data_27.json", "valid": "true"},
		{"phase": "format validation", "test": "uri format is invalid", "schema": "format/schema_6.json", "data": "format/data_28.json", "valid": "false", "errors": "format"},
		{"phase": "format validation", "test": "uri format is invalid", "schema": "format/schema_6.json", "data": "format/data_13.json", "valid": "false", "errors": "format"},
	}

	//TODO Pass failed tests : id(s) as scope for references is not implemented yet
	//map[string]string{"phase": "change resolution scope", "test": "changed scope ref valid", "schema": "refRemote/schema_3.json", "data": "refRemote/data_30.json", "valid": "true"},
	//map[string]string{"phase": "change resolution scope", "test": "changed scope ref invalid", "schema": "refRemote/schema_3.json", "data": "refRemote/data_31.json", "valid": "false"}}

	// Setup a small http server on localhost:1234 for testing purposes

	wd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}

	testwd := wd + "/json_schema_test_suite"

	go func() {
		err := http.ListenAndServe(":1234", http.FileServer(http.Dir(testwd+"/refRemote/remoteFiles")))
		if err != nil {
			panic(err.Error())
		}
	}()

	// Used for remote schema in ref/schema_5.json that defines "regex" type
	FormatCheckers.Add("regex", alwaysTrueFormatChecker{})

	// Custom Formatter
	FormatCheckers.Add("invoice", invoiceFormatChecker{})

	// Launch tests

	for testJsonIndex, testJson := range JsonSchemaTestSuiteMap {

		fmt.Printf("Test (%d) | %s :: %s\n", testJsonIndex, testJson["phase"], testJson["test"])

		schemaLoader := NewReferenceLoader("file://" + testwd + "/" + testJson["schema"])
		documentLoader := NewReferenceLoader("file://" + testwd + "/" + testJson["data"])

		// validate
		result, err := Validate(schemaLoader, documentLoader)
		if err != nil {
			t.Errorf("Error (%s)\n", err.Error())
		}
		givenValid := result.Valid()

		if displayErrorMessages {
			for vErrI, vErr := range result.Errors() {
				fmt.Printf("  Error (%d) | %s\n", vErrI, vErr)
			}
		}

		expectedValid, _ := strconv.ParseBool(testJson["valid"])
		if givenValid != expectedValid {
			t.Errorf("Test failed : %s :: %s, expects %t, given %t\n", testJson["phase"], testJson["test"], expectedValid, givenValid)
		}

		if !givenValid && testJson["errors"] != "" {
			expectedErrors := rxSplitErrors.Split(testJson["errors"], -1)
			errors := result.Errors()
			if len(errors) != len(expectedErrors) {
				t.Errorf("Test failed : %s :: %s, expects %d errors, given %d errors\n", testJson["phase"], testJson["test"], len(expectedErrors), len(errors))
			}

			actualErrors := make([]string, 0)
			for _, e := range result.Errors() {
				actualErrors = append(actualErrors, e.Type())
			}

			sort.Strings(actualErrors)
			sort.Strings(expectedErrors)
			if !reflect.DeepEqual(actualErrors, expectedErrors) {
				t.Errorf("Test failed : %s :: %s, expected '%s' errors, given '%s' errors\n", testJson["phase"], testJson["test"], strings.Join(expectedErrors, ", "), strings.Join(actualErrors, ", "))
			}
		}

	}

	fmt.Printf("\n%d tests performed / %d total tests to perform ( %.2f %% )\n", len(JsonSchemaTestSuiteMap), 248, float32(len(JsonSchemaTestSuiteMap))/248.0*100.0)
}

const circularReference = `{
	"type": "object",
	"properties": {
		"games": {
			"type": "array",
			"items": {
				"$ref": "#/definitions/game"
			}
		}
	},
	"definitions": {
		"game": {
			"type": "object",
			"properties": {
				"winner": {
					"$ref": "#/definitions/player"
				},
				"loser": {
					"$ref": "#/definitions/player"
				}
			}
		},
		"player": {
			"type": "object",
			"properties": {
				"user": {
					"$ref": "#/definitions/user"
				},
				"game": {
					"$ref": "#/definitions/game"
				}
			}
		},
		"user": {
			"type": "object",
			"properties": {
				"fullName": {
					"type": "string"
				}
			}
		}
	}
}`

func TestCircularReference(t *testing.T) {
	loader := NewStringLoader(circularReference)
	// call the target function
	_, err := NewSchema(loader)
	if err != nil {
		t.Errorf("Got error: %s", err.Error())
	}
}

// From http://json-schema.org/examples.html
const simpleSchema = `{
  "title": "Example Schema",
  "type": "object",
  "properties": {
    "firstName": {
      "type": "string"
    },
    "lastName": {
      "type": "string"
    },
    "age": {
      "description": "Age in years",
      "type": "integer",
      "minimum": 0
    }
  },
  "required": ["firstName", "lastName"]
}`

func TestLoaders(t *testing.T) {
	// setup reader loader
	reader := bytes.NewBufferString(simpleSchema)
	readerLoader, wrappedReader := NewReaderLoader(reader)

	// drain reader
	by, err := ioutil.ReadAll(wrappedReader)
	assert.Nil(t, err)
	assert.Equal(t, simpleSchema, string(by))

	// setup writer loaders
	writer := &bytes.Buffer{}
	writerLoader, wrappedWriter := NewWriterLoader(writer)

	// fill writer
	n, err := io.WriteString(wrappedWriter, simpleSchema)
	assert.Nil(t, err)
	assert.Equal(t, n, len(simpleSchema))

	loaders := []JSONLoader{
		NewStringLoader(simpleSchema),
		readerLoader,
		writerLoader,
	}

	for _, l := range loaders {
		_, err := NewSchema(l)
		assert.Nil(t, err, "loader: %T", l)
	}
}

const invalidPattern = `{
  "title": "Example Pattern",
  "type": "object",
  "properties": {
    "invalid": {
      "type": "string",
      "pattern": 99999
    }
  }
}`

func TestLoadersWithInvalidPattern(t *testing.T) {
	// setup reader loader
	reader := bytes.NewBufferString(invalidPattern)
	readerLoader, wrappedReader := NewReaderLoader(reader)

	// drain reader
	by, err := ioutil.ReadAll(wrappedReader)
	assert.Nil(t, err)
	assert.Equal(t, invalidPattern, string(by))

	// setup writer loaders
	writer := &bytes.Buffer{}
	writerLoader, wrappedWriter := NewWriterLoader(writer)

	// fill writer
	n, err := io.WriteString(wrappedWriter, invalidPattern)
	assert.Nil(t, err)
	assert.Equal(t, n, len(invalidPattern))

	loaders := []JSONLoader{
		NewStringLoader(invalidPattern),
		readerLoader,
		writerLoader,
	}

	for _, l := range loaders {
		_, err := NewSchema(l)
		assert.NotNil(t, err, "expected error loading invalid pattern: %T", l)
	}
}
