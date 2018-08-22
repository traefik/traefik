package data

// The following code is a modified copy of functions found in strcase, found here: https://github.com/iancoleman/strcase

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

// ToCamel converts a string to CamelCase
func ToCamel(s string) string {
	s = addWordBoundariesToNumbers(s)
	s = strings.TrimSpace(s)
	b := bytes.NewBuffer(make([]byte, 0))
	c := true
	for _, v := range s {
		if v >= 'A' && v <= 'Z' {
			b.WriteString(string(v))
		}
		if v >= '0' && v <= '9' {
			b.WriteString(string(v))
		}
		if v >= 'a' && v <= 'z' {
			if c {
				b.WriteString(string(unicode.ToUpper(v)))
			} else {
				b.WriteString(string(v))
			}
		}
		if v == '_' || v == ' ' || v == '-' {
			c = true
		} else {
			c = false
		}
	}
	return strings.TrimSpace(b.String())
}

var numberSequence = regexp.MustCompile(`([a-zA-Z])(\d+)([a-zA-Z]?)`)
var numberReplacement = []byte(`$1 $2 $3`)

func addWordBoundariesToNumbers(s string) string {
	b := []byte(s)
	b = numberSequence.ReplaceAll(b, numberReplacement)
	return string(b)
}
