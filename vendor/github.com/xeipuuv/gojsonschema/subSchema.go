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
// description      Defines the structure of a sub-subSchema.
//                  A sub-subSchema can contain other sub-schemas.
//
// created          27-02-2013

package gojsonschema

import (
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/xeipuuv/gojsonreference"
)

const (
	KEY_SCHEMA                = "$schema"
	KEY_ID                    = "id"
	KEY_ID_NEW                = "$id"
	KEY_REF                   = "$ref"
	KEY_TITLE                 = "title"
	KEY_DESCRIPTION           = "description"
	KEY_TYPE                  = "type"
	KEY_ITEMS                 = "items"
	KEY_ADDITIONAL_ITEMS      = "additionalItems"
	KEY_PROPERTIES            = "properties"
	KEY_PATTERN_PROPERTIES    = "patternProperties"
	KEY_ADDITIONAL_PROPERTIES = "additionalProperties"
	KEY_PROPERTY_NAMES        = "propertyNames"
	KEY_DEFINITIONS           = "definitions"
	KEY_MULTIPLE_OF           = "multipleOf"
	KEY_MINIMUM               = "minimum"
	KEY_MAXIMUM               = "maximum"
	KEY_EXCLUSIVE_MINIMUM     = "exclusiveMinimum"
	KEY_EXCLUSIVE_MAXIMUM     = "exclusiveMaximum"
	KEY_MIN_LENGTH            = "minLength"
	KEY_MAX_LENGTH            = "maxLength"
	KEY_PATTERN               = "pattern"
	KEY_FORMAT                = "format"
	KEY_MIN_PROPERTIES        = "minProperties"
	KEY_MAX_PROPERTIES        = "maxProperties"
	KEY_DEPENDENCIES          = "dependencies"
	KEY_REQUIRED              = "required"
	KEY_MIN_ITEMS             = "minItems"
	KEY_MAX_ITEMS             = "maxItems"
	KEY_UNIQUE_ITEMS          = "uniqueItems"
	KEY_CONTAINS              = "contains"
	KEY_CONST                 = "const"
	KEY_ENUM                  = "enum"
	KEY_ONE_OF                = "oneOf"
	KEY_ANY_OF                = "anyOf"
	KEY_ALL_OF                = "allOf"
	KEY_NOT                   = "not"
	KEY_IF                    = "if"
	KEY_THEN                  = "then"
	KEY_ELSE                  = "else"
)

type subSchema struct {
	draft *Draft

	// basic subSchema meta properties
	id          *gojsonreference.JsonReference
	title       *string
	description *string

	property string

	// Types associated with the subSchema
	types jsonSchemaType

	// Reference url
	ref *gojsonreference.JsonReference
	// Schema referenced
	refSchema *subSchema

	// hierarchy
	parent                      *subSchema
	itemsChildren               []*subSchema
	itemsChildrenIsSingleSchema bool
	propertiesChildren          []*subSchema

	// validation : number / integer
	multipleOf       *big.Rat
	maximum          *big.Rat
	exclusiveMaximum *big.Rat
	minimum          *big.Rat
	exclusiveMinimum *big.Rat

	// validation : string
	minLength *int
	maxLength *int
	pattern   *regexp.Regexp
	format    string

	// validation : object
	minProperties *int
	maxProperties *int
	required      []string

	dependencies         map[string]interface{}
	additionalProperties interface{}
	patternProperties    map[string]*subSchema
	propertyNames        *subSchema

	// validation : array
	minItems    *int
	maxItems    *int
	uniqueItems bool
	contains    *subSchema

	additionalItems interface{}

	// validation : all
	_const *string //const is a golang keyword
	enum   []string

	// validation : subSchema
	oneOf []*subSchema
	anyOf []*subSchema
	allOf []*subSchema
	not   *subSchema
	_if   *subSchema // if/else are golang keywords
	_then *subSchema
	_else *subSchema
}

func (s *subSchema) AddConst(i interface{}) error {

	is, err := marshalWithoutNumber(i)
	if err != nil {
		return err
	}
	s._const = is
	return nil
}

func (s *subSchema) AddEnum(i interface{}) error {

	is, err := marshalWithoutNumber(i)
	if err != nil {
		return err
	}

	if isStringInSlice(s.enum, *is) {
		return errors.New(formatErrorDescription(
			Locale.KeyItemsMustBeUnique(),
			ErrorDetails{"key": KEY_ENUM},
		))
	}

	s.enum = append(s.enum, *is)

	return nil
}

func (s *subSchema) ContainsEnum(i interface{}) (bool, error) {

	is, err := marshalWithoutNumber(i)
	if err != nil {
		return false, err
	}

	return isStringInSlice(s.enum, *is), nil
}

func (s *subSchema) AddOneOf(subSchema *subSchema) {
	s.oneOf = append(s.oneOf, subSchema)
}

func (s *subSchema) AddAllOf(subSchema *subSchema) {
	s.allOf = append(s.allOf, subSchema)
}

func (s *subSchema) AddAnyOf(subSchema *subSchema) {
	s.anyOf = append(s.anyOf, subSchema)
}

func (s *subSchema) SetNot(subSchema *subSchema) {
	s.not = subSchema
}

func (s *subSchema) SetIf(subSchema *subSchema) {
	s._if = subSchema
}

func (s *subSchema) SetThen(subSchema *subSchema) {
	s._then = subSchema
}

func (s *subSchema) SetElse(subSchema *subSchema) {
	s._else = subSchema
}

func (s *subSchema) AddRequired(value string) error {

	if isStringInSlice(s.required, value) {
		return errors.New(formatErrorDescription(
			Locale.KeyItemsMustBeUnique(),
			ErrorDetails{"key": KEY_REQUIRED},
		))
	}

	s.required = append(s.required, value)

	return nil
}

func (s *subSchema) AddItemsChild(child *subSchema) {
	s.itemsChildren = append(s.itemsChildren, child)
}

func (s *subSchema) AddPropertiesChild(child *subSchema) {
	s.propertiesChildren = append(s.propertiesChildren, child)
}

func (s *subSchema) PatternPropertiesString() string {

	if s.patternProperties == nil || len(s.patternProperties) == 0 {
		return STRING_UNDEFINED // should never happen
	}

	patternPropertiesKeySlice := []string{}
	for pk := range s.patternProperties {
		patternPropertiesKeySlice = append(patternPropertiesKeySlice, `"`+pk+`"`)
	}

	if len(patternPropertiesKeySlice) == 1 {
		return patternPropertiesKeySlice[0]
	}

	return "[" + strings.Join(patternPropertiesKeySlice, ",") + "]"

}
