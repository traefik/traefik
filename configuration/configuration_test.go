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
			value: "Name:foo TLS:goo TLS CA:car Redirect.EntryPoint:RedirectEntryPoint Redirect.Regex:RedirectRegex Redirect.Replacement:RedirectReplacement Compress:true WhiteListSourceRange:WhiteListSourceRange ProxyProtocol.TrustedIPs:192.168.0.1 Address::8000",
			expectedResult: map[string]string{
				"name":                     "foo",
				"address":                  ":8000",
				"ca":                       "car",
				"tls":                      "goo",
				"tls_acme":                 "TLS",
				"redirect_entrypoint":      "RedirectEntryPoint",
				"redirect_regex":           "RedirectRegex",
				"redirect_replacement":     "RedirectReplacement",
				"whitelistsourcerange":     "WhiteListSourceRange",
				"proxyprotocol_trustedips": "192.168.0.1",
				"compress":                 "true",
			},
		},
		{
			name:  "compress on",
			value: "name:foo Compress:on",
			expectedResult: map[string]string{
				"name":     "foo",
				"compress": "on",
			},
		},
		{
			name:  "TLS",
			value: "Name:foo TLS:goo TLS",
			expectedResult: map[string]string{
				"name":     "foo",
				"tls":      "goo",
				"tls_acme": "TLS",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			conf := parseEntryPointsConfiguration(test.value)

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
			name:                   "all parameters camelcase",
			expression:             "Name:foo Address::8000 TLS:goo,gii TLS CA:car Redirect.EntryPoint:RedirectEntryPoint Redirect.Regex:RedirectRegex Redirect.Replacement:RedirectReplacement Compress:true WhiteListSourceRange:Range ProxyProtocol.TrustedIPs:192.168.0.1",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				Address: ":8000",
				Redirect: &Redirect{
					EntryPoint:  "RedirectEntryPoint",
					Regex:       "RedirectRegex",
					Replacement: "RedirectReplacement",
				},
				Compress: true,
				ProxyProtocol: &ProxyProtocol{
					TrustedIPs: []string{"192.168.0.1"},
				},
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
			name:                   "all parameters lowercase",
			expression:             "name:foo address::8000 tls:goo,gii tls ca:car redirect.entryPoint:RedirectEntryPoint redirect.regex:RedirectRegex redirect.replacement:RedirectReplacement compress:true whiteListSourceRange:Range proxyProtocol.TrustedIPs:192.168.0.1",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				Address: ":8000",
				Redirect: &Redirect{
					EntryPoint:  "RedirectEntryPoint",
					Regex:       "RedirectRegex",
					Replacement: "RedirectReplacement",
				},
				Compress: true,
				ProxyProtocol: &ProxyProtocol{
					TrustedIPs: []string{"192.168.0.1"},
				},
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
