package configuration

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultConfigFile = "traefik.toml"

func Test_parseEntryPointsConfiguration(t *testing.T) {
	testCases := []struct {
		name           string
		value          string
		expectedResult map[string]string
	}{
		{
			name:  "all parameters",
			value: "Name:foo TLS:goo TLS CA:car Redirect.EntryPoint:RedirectEntryPoint Redirect.Regex:RedirectRegex Redirect.Replacement:RedirectReplacement Compress:true WhiteListSourceRange:WhiteListSourceRange ProxyProtocol.TrustedIPs:192.168.0.1 ProxyProtocol.Insecure:false Address::8000",
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
				"proxyprotocol_insecure":   "false",
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
			expression:             "Name:foo Address::8000 TLS:goo,gii TLS CA:car Redirect.EntryPoint:RedirectEntryPoint Redirect.Regex:RedirectRegex Redirect.Replacement:RedirectReplacement Compress:true WhiteListSourceRange:Range ProxyProtocol.TrustedIPs:192.168.0.1 ForwardedHeaders.TrustedIPs:10.0.0.3/24,20.0.0.3/24",
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
				ForwardedHeaders: &ForwardedHeaders{
					TrustedIPs: []string{"10.0.0.3/24", "20.0.0.3/24"},
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
			expression:             "name:foo address::8000 tls:goo,gii tls ca:car redirect.entryPoint:RedirectEntryPoint redirect.regex:RedirectRegex redirect.replacement:RedirectReplacement compress:true whiteListSourceRange:Range proxyProtocol.trustedIPs:192.168.0.1 forwardedHeaders.trustedIPs:10.0.0.3/24,20.0.0.3/24",
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
				ForwardedHeaders: &ForwardedHeaders{
					TrustedIPs: []string{"10.0.0.3/24", "20.0.0.3/24"},
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
			name:                   "default",
			expression:             "Name:foo",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				WhitelistSourceRange: []string{},
				ForwardedHeaders:     &ForwardedHeaders{Insecure: true},
			},
		},
		{
			name:                   "ForwardedHeaders insecure true",
			expression:             "Name:foo ForwardedHeaders.Insecure:true",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				WhitelistSourceRange: []string{},
				ForwardedHeaders:     &ForwardedHeaders{Insecure: true},
			},
		},
		{
			name:                   "ForwardedHeaders insecure false",
			expression:             "Name:foo ForwardedHeaders.Insecure:false",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				WhitelistSourceRange: []string{},
				ForwardedHeaders:     &ForwardedHeaders{Insecure: false},
			},
		},
		{
			name:                   "ForwardedHeaders TrustedIPs",
			expression:             "Name:foo ForwardedHeaders.TrustedIPs:10.0.0.3/24,20.0.0.3/24",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				WhitelistSourceRange: []string{},
				ForwardedHeaders: &ForwardedHeaders{
					TrustedIPs: []string{"10.0.0.3/24", "20.0.0.3/24"},
				},
			},
		},
		{
			name:                   "ProxyProtocol insecure true",
			expression:             "Name:foo ProxyProtocol.Insecure:true",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				WhitelistSourceRange: []string{},
				ForwardedHeaders:     &ForwardedHeaders{Insecure: true},
				ProxyProtocol:        &ProxyProtocol{Insecure: true},
			},
		},
		{
			name:                   "ProxyProtocol insecure false",
			expression:             "Name:foo ProxyProtocol.Insecure:false",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				WhitelistSourceRange: []string{},
				ForwardedHeaders:     &ForwardedHeaders{Insecure: true},
				ProxyProtocol:        &ProxyProtocol{},
			},
		},
		{
			name:                   "ProxyProtocol TrustedIPs",
			expression:             "Name:foo ProxyProtocol.TrustedIPs:10.0.0.3/24,20.0.0.3/24",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				WhitelistSourceRange: []string{},
				ForwardedHeaders:     &ForwardedHeaders{Insecure: true},
				ProxyProtocol: &ProxyProtocol{
					TrustedIPs: []string{"10.0.0.3/24", "20.0.0.3/24"},
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
				ForwardedHeaders:     &ForwardedHeaders{Insecure: true},
			},
		},
		{
			name:                   "compress true",
			expression:             "Name:foo Compress:true",
			expectedEntryPointName: "foo",
			expectedEntryPoint: &EntryPoint{
				Compress:             true,
				WhitelistSourceRange: []string{},
				ForwardedHeaders:     &ForwardedHeaders{Insecure: true},
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

func TestSetEffectiveConfigurationGraceTimeout(t *testing.T) {
	tests := []struct {
		desc                  string
		legacyGraceTimeout    time.Duration
		lifeCycleGraceTimeout time.Duration
		wantGraceTimeout      time.Duration
	}{
		{
			desc:               "legacy grace timeout given only",
			legacyGraceTimeout: 5 * time.Second,
			wantGraceTimeout:   5 * time.Second,
		},
		{
			desc:                  "legacy and life cycle grace timeouts given",
			legacyGraceTimeout:    5 * time.Second,
			lifeCycleGraceTimeout: 12 * time.Second,
			wantGraceTimeout:      5 * time.Second,
		},
		{
			desc:                  "legacy grace timeout omitted",
			legacyGraceTimeout:    0,
			lifeCycleGraceTimeout: 12 * time.Second,
			wantGraceTimeout:      12 * time.Second,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			gc := &GlobalConfiguration{
				GraceTimeOut: flaeg.Duration(test.legacyGraceTimeout),
			}
			if test.lifeCycleGraceTimeout > 0 {
				gc.LifeCycle = &LifeCycle{
					GraceTimeOut: flaeg.Duration(test.lifeCycleGraceTimeout),
				}
			}

			gc.SetEffectiveConfiguration(defaultConfigFile)

			gotGraceTimeout := time.Duration(gc.LifeCycle.GraceTimeOut)
			if gotGraceTimeout != test.wantGraceTimeout {
				t.Fatalf("got effective grace timeout %d, want %d", gotGraceTimeout, test.wantGraceTimeout)
			}

		})
	}
}

func TestSetEffectiveConfigurationFileProviderFilename(t *testing.T) {
	tests := []struct {
		desc                     string
		fileProvider             *file.Provider
		wantFileProviderFilename string
	}{
		{
			desc:                     "no filename for file provider given",
			fileProvider:             &file.Provider{},
			wantFileProviderFilename: defaultConfigFile,
		},
		{
			desc:                     "filename for file provider given",
			fileProvider:             &file.Provider{BaseProvider: provider.BaseProvider{Filename: "other.toml"}},
			wantFileProviderFilename: "other.toml",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			gc := &GlobalConfiguration{
				File: test.fileProvider,
			}

			gc.SetEffectiveConfiguration(defaultConfigFile)

			gotFileProviderFilename := gc.File.Filename
			if gotFileProviderFilename != test.wantFileProviderFilename {
				t.Fatalf("got file provider file name %q, want %q", gotFileProviderFilename, test.wantFileProviderFilename)
			}
		})
	}
}
