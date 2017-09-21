package configuration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseEntryPointsConfiguration(t *testing.T) {
	testCases := []struct {
		name           string
		value          string
		expectedResult map[string]string
	}{
		{
			name:  "all parameters",
			value: "Name:foo Address:bar TLS:goo TLS CA:car Redirect.EntryPoint:RedirectEntryPoint Redirect.Regex:RedirectRegex Redirect.Replacement:RedirectReplacement Compress:true WhiteListSourceRange:WhiteListSourceRange ProxyProtocol:true",
			expectedResult: map[string]string{
				"Name":                 "foo",
				"Address":              "bar",
				"CA":                   "car",
				"TLS":                  "goo",
				"TLSACME":              "TLS",
				"RedirectEntryPoint":   "RedirectEntryPoint",
				"RedirectRegex":        "RedirectRegex",
				"RedirectReplacement":  "RedirectReplacement",
				"WhiteListSourceRange": "WhiteListSourceRange",
				"ProxyProtocol":        "true",
				"Compress":             "true",
			},
		},
		{
			name:  "proxy protocol on",
			value: "Name:foo ProxyProtocol:on",
			expectedResult: map[string]string{
				"Name":          "foo",
				"ProxyProtocol": "on",
			},
		},
		{
			name:  "compress on",
			value: "Name:foo Compress:on",
			expectedResult: map[string]string{
				"Name":     "foo",
				"Compress": "on",
			},
		},
		{
			name:  "TLS",
			value: "Name:foo TLS:goo TLS",
			expectedResult: map[string]string{
				"Name":    "foo",
				"TLS":     "goo",
				"TLSACME": "TLS",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			conf, err := parseEntryPointsConfiguration(test.value)
			if err != nil {
				t.Error(err)
			}

			for key, value := range conf {
				fmt.Println(key, value)
			}

			assert.Len(t, conf, len(test.expectedResult))
			assert.Equal(t, test.expectedResult, conf)
		})
	}
}

func Test_toBool(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		key          string
		expectedBool bool
	}{
		{
			name:         "on",
			value:        "on",
			key:          "foo",
			expectedBool: true,
		},
		{
			name:         "true",
			value:        "true",
			key:          "foo",
			expectedBool: true,
		},
		{
			name:         "enable",
			value:        "enable",
			key:          "foo",
			expectedBool: true,
		},
		{
			name:         "arbitrary string",
			value:        "bar",
			key:          "foo",
			expectedBool: false,
		},
		{
			name:         "no existing entry",
			value:        "bar",
			key:          "fii",
			expectedBool: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			conf := map[string]string{
				"foo": test.value,
			}

			result := toBool(conf, test.key)

			assert.Equal(t, test.expectedBool, result)
		})
	}
}

func TestEntryPoints_Set(t *testing.T) {
	testCases := []struct {
		name                   string
		expression             string
		expectedEntryPointName string
		expectedEntryPoint     *EntryPoint
	}{
		{
			name:                   "all parameters",
			expression:             "Name:foo Address:bar TLS:goo,gii TLS CA:car Redirect.EntryPoint:RedirectEntryPoint Redirect.Regex:RedirectRegex Redirect.Replacement:RedirectReplacement Compress:true WhiteListSourceRange:Range ProxyProtocol:true",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				Address: "bar",
				Redirect: &Redirect{
					EntryPoint:  "RedirectEntryPoint",
					Regex:       "RedirectRegex",
					Replacement: "RedirectReplacement",
				},
				Compress:             true,
				ProxyProtocol:        true,
				WhitelistSourceRange: []string{"Range"},
				TLS: &TLS{
					ClientCAFiles: []string{"car"},
					Certificates: Certificates{
						{
							CertFile: FileOrContent("goo"),
							KeyFile:  FileOrContent("gii"),
						},
					},
				},
			},
		},
		{
			name:                   "compress on",
			expression:             "Name:foo Compress:on",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				Compress:             true,
				WhitelistSourceRange: []string{},
			},
		},
		{
			name:                   "compress true",
			expression:             "Name:foo Compress:true",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				Compress:             true,
				WhitelistSourceRange: []string{},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			eps := EntryPoints{}
			err := eps.Set(test.expression)
			require.NoError(t, err)

			ep := eps[test.expectedEntryPointName]
			assert.EqualValues(t, test.expectedEntryPoint, ep)
		})
	}
}
