package v2

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/docker/distribution/reference"
)

type urlBuilderTestCase struct {
	description  string
	expectedPath string
	build        func() (string, error)
}

func makeURLBuilderTestCases(urlBuilder *URLBuilder) []urlBuilderTestCase {
	fooBarRef, _ := reference.WithName("foo/bar")
	return []urlBuilderTestCase{
		{
			description:  "test base url",
			expectedPath: "/v2/",
			build:        urlBuilder.BuildBaseURL,
		},
		{
			description:  "test tags url",
			expectedPath: "/v2/foo/bar/tags/list",
			build: func() (string, error) {
				return urlBuilder.BuildTagsURL(fooBarRef)
			},
		},
		{
			description:  "test manifest url",
			expectedPath: "/v2/foo/bar/manifests/tag",
			build: func() (string, error) {
				ref, _ := reference.WithTag(fooBarRef, "tag")
				return urlBuilder.BuildManifestURL(ref)
			},
		},
		{
			description:  "build blob url",
			expectedPath: "/v2/foo/bar/blobs/sha256:3b3692957d439ac1928219a83fac91e7bf96c153725526874673ae1f2023f8d5",
			build: func() (string, error) {
				ref, _ := reference.WithDigest(fooBarRef, "sha256:3b3692957d439ac1928219a83fac91e7bf96c153725526874673ae1f2023f8d5")
				return urlBuilder.BuildBlobURL(ref)
			},
		},
		{
			description:  "build blob upload url",
			expectedPath: "/v2/foo/bar/blobs/uploads/",
			build: func() (string, error) {
				return urlBuilder.BuildBlobUploadURL(fooBarRef)
			},
		},
		{
			description:  "build blob upload url with digest and size",
			expectedPath: "/v2/foo/bar/blobs/uploads/?digest=sha256%3A3b3692957d439ac1928219a83fac91e7bf96c153725526874673ae1f2023f8d5&size=10000",
			build: func() (string, error) {
				return urlBuilder.BuildBlobUploadURL(fooBarRef, url.Values{
					"size":   []string{"10000"},
					"digest": []string{"sha256:3b3692957d439ac1928219a83fac91e7bf96c153725526874673ae1f2023f8d5"},
				})
			},
		},
		{
			description:  "build blob upload chunk url",
			expectedPath: "/v2/foo/bar/blobs/uploads/uuid-part",
			build: func() (string, error) {
				return urlBuilder.BuildBlobUploadChunkURL(fooBarRef, "uuid-part")
			},
		},
		{
			description:  "build blob upload chunk url with digest and size",
			expectedPath: "/v2/foo/bar/blobs/uploads/uuid-part?digest=sha256%3A3b3692957d439ac1928219a83fac91e7bf96c153725526874673ae1f2023f8d5&size=10000",
			build: func() (string, error) {
				return urlBuilder.BuildBlobUploadChunkURL(fooBarRef, "uuid-part", url.Values{
					"size":   []string{"10000"},
					"digest": []string{"sha256:3b3692957d439ac1928219a83fac91e7bf96c153725526874673ae1f2023f8d5"},
				})
			},
		},
	}
}

// TestURLBuilder tests the various url building functions, ensuring they are
// returning the expected values.
func TestURLBuilder(t *testing.T) {
	roots := []string{
		"http://example.com",
		"https://example.com",
		"http://localhost:5000",
		"https://localhost:5443",
	}

	doTest := func(relative bool) {
		for _, root := range roots {
			urlBuilder, err := NewURLBuilderFromString(root, relative)
			if err != nil {
				t.Fatalf("unexpected error creating urlbuilder: %v", err)
			}

			for _, testCase := range makeURLBuilderTestCases(urlBuilder) {
				url, err := testCase.build()
				if err != nil {
					t.Fatalf("%s: error building url: %v", testCase.description, err)
				}
				expectedURL := testCase.expectedPath
				if !relative {
					expectedURL = root + expectedURL
				}

				if url != expectedURL {
					t.Fatalf("%s: %q != %q", testCase.description, url, expectedURL)
				}
			}
		}
	}
	doTest(true)
	doTest(false)
}

func TestURLBuilderWithPrefix(t *testing.T) {
	roots := []string{
		"http://example.com/prefix/",
		"https://example.com/prefix/",
		"http://localhost:5000/prefix/",
		"https://localhost:5443/prefix/",
	}

	doTest := func(relative bool) {
		for _, root := range roots {
			urlBuilder, err := NewURLBuilderFromString(root, relative)
			if err != nil {
				t.Fatalf("unexpected error creating urlbuilder: %v", err)
			}

			for _, testCase := range makeURLBuilderTestCases(urlBuilder) {
				url, err := testCase.build()
				if err != nil {
					t.Fatalf("%s: error building url: %v", testCase.description, err)
				}

				expectedURL := testCase.expectedPath
				if !relative {
					expectedURL = root[0:len(root)-1] + expectedURL
				}
				if url != expectedURL {
					t.Fatalf("%s: %q != %q", testCase.description, url, expectedURL)
				}
			}
		}
	}
	doTest(true)
	doTest(false)
}

type builderFromRequestTestCase struct {
	request *http.Request
	base    string
}

func TestBuilderFromRequest(t *testing.T) {
	u, err := url.Parse("http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	testRequests := []struct {
		name       string
		request    *http.Request
		base       string
		configHost url.URL
	}{
		{
			name:    "no forwarded header",
			request: &http.Request{URL: u, Host: u.Host},
			base:    "http://example.com",
		},
		{
			name: "https protocol forwarded with a non-standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Proto": []string{"https"},
			}},
			base: "http://example.com",
		},
		{
			name: "forwarded protocol is the same",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Proto": []string{"https"},
			}},
			base: "https://example.com",
		},
		{
			name: "forwarded host with a non-standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Host": []string{"first.example.com"},
			}},
			base: "http://first.example.com",
		},
		{
			name: "forwarded multiple hosts a with non-standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Host": []string{"first.example.com, proxy1.example.com"},
			}},
			base: "http://first.example.com",
		},
		{
			name: "host configured in config file takes priority",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Host": []string{"first.example.com, proxy1.example.com"},
			}},
			base: "https://third.example.com:5000",
			configHost: url.URL{
				Scheme: "https",
				Host:   "third.example.com:5000",
			},
		},
		{
			name: "forwarded host and port with just one non-standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Host": []string{"first.example.com:443"},
			}},
			base: "http://first.example.com:443",
		},
		{
			name: "forwarded port with a non-standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Port": []string{"5000"},
			}},
			base: "http://example.com:5000",
		},
		{
			name: "forwarded multiple ports with a non-standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Port": []string{"443 , 5001"},
			}},
			base: "http://example.com:443",
		},
		{
			name: "several non-standard headers",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Proto": []string{"https"},
				"X-Forwarded-Host":  []string{" first.example.com "},
				"X-Forwarded-Port":  []string{" 12345 \t"},
			}},
			base: "http://first.example.com:12345",
		},
		{
			name: "forwarded host with port supplied takes priority",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Host": []string{"first.example.com:5000"},
				"X-Forwarded-Port": []string{"80"},
			}},
			base: "http://first.example.com:5000",
		},
		{
			name: "malformed forwarded port",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Host": []string{"first.example.com"},
				"X-Forwarded-Port": []string{"abcd"},
			}},
			base: "http://first.example.com",
		},
		{
			name: "forwarded protocol and addr using standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded": []string{`proto=https;for="192.168.22.30:80"`},
			}},
			base: "https://192.168.22.30:80",
		},
		{
			name: "forwarded addr takes priority over host",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded": []string{`host=reg.example.com;for="192.168.22.30:5000"`},
			}},
			base: "http://192.168.22.30:5000",
		},
		{
			name: "forwarded host and protocol using standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded": []string{`host=reg.example.com;proto=https`},
			}},
			base: "https://reg.example.com",
		},
		{
			name: "process just the first standard forwarded header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded": []string{`host="reg.example.com:88";proto=http`, `host=reg.example.com;proto=https`},
			}},
			base: "http://reg.example.com:88",
		},
		{
			name: "process just the first list element of standard header",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded": []string{`for="reg.example.com:443";proto=https, for="reg.example.com:80";proto=http`},
			}},
			base: "https://reg.example.com:443",
		},
		{
			name: "IPv6 address override port",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded":        []string{`for="2607:f0d0:1002:51::4"`},
				"X-Forwarded-Port": []string{"5001"},
			}},
			base: "http://[2607:f0d0:1002:51::4]:5001",
		},
		{
			name: "IPv6 address with port",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded":        []string{`for="[2607:f0d0:1002:51::4]:4000"`},
				"X-Forwarded-Port": []string{"5001"},
			}},
			base: "http://[2607:f0d0:1002:51::4]:4000",
		},
		{
			name: "IPv6 long address override port",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded":        []string{`for="2607:f0d0:1002:0051:0000:0000:0000:0004"`},
				"X-Forwarded-Port": []string{"5001"},
			}},
			base: "http://[2607:f0d0:1002:0051:0000:0000:0000:0004]:5001",
		},
		{
			name: "IPv6 long address enclosed in brackets - be benevolent",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded":        []string{`for="[2607:f0d0:1002:0051:0000:0000:0000:0004]"`},
				"X-Forwarded-Port": []string{"5001"},
			}},
			base: "http://[2607:f0d0:1002:0051:0000:0000:0000:0004]:5001",
		},
		{
			name: "IPv6 long address with port",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded":        []string{`for="[2607:f0d0:1002:0051:0000:0000:0000:0004]:4321"`},
				"X-Forwarded-Port": []string{"5001"},
			}},
			base: "http://[2607:f0d0:1002:0051:0000:0000:0000:0004]:4321",
		},
		{
			name: "IPv6 address with zone ID",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded":        []string{`for="fe80::bd0f:a8bc:6480:238b%11"`},
				"X-Forwarded-Port": []string{"5001"},
			}},
			base: "http://[fe80::bd0f:a8bc:6480:238b%2511]:5001",
		},
		{
			name: "IPv6 address with zone ID and port",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded":        []string{`for="[fe80::bd0f:a8bc:6480:238b%eth0]:12345"`},
				"X-Forwarded-Port": []string{"5001"},
			}},
			base: "http://[fe80::bd0f:a8bc:6480:238b%25eth0]:12345",
		},
		{
			name: "IPv6 address without port",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"Forwarded": []string{`for="::FFFF:129.144.52.38"`},
			}},
			base: "http://[::FFFF:129.144.52.38]",
		},
		{
			name: "non-standard and standard forward headers",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Proto": []string{`https`},
				"X-Forwarded-Host":  []string{`first.example.com`},
				"X-Forwarded-Port":  []string{``},
				"Forwarded":         []string{`host=first.example.com; proto=https`},
			}},
			base: "https://first.example.com",
		},
		{
			name: "non-standard headers take precedence over standard one",
			request: &http.Request{URL: u, Host: u.Host, Header: http.Header{
				"X-Forwarded-Proto": []string{`http`},
				"Forwarded":         []string{`host=second.example.com; proto=https`},
				"X-Forwarded-Host":  []string{`first.example.com`},
				"X-Forwarded-Port":  []string{`4000`},
			}},
			base: "http://first.example.com:4000",
		},
	}

	doTest := func(relative bool) {
		for _, tr := range testRequests {
			var builder *URLBuilder
			if tr.configHost.Scheme != "" && tr.configHost.Host != "" {
				builder = NewURLBuilder(&tr.configHost, relative)
			} else {
				builder = NewURLBuilderFromRequest(tr.request, relative)
			}

			for _, testCase := range makeURLBuilderTestCases(builder) {
				buildURL, err := testCase.build()
				if err != nil {
					t.Fatalf("[relative=%t, request=%q, case=%q]: error building url: %v", relative, tr.name, testCase.description, err)
				}

				var expectedURL string
				proto, ok := tr.request.Header["X-Forwarded-Proto"]
				if !ok {
					expectedURL = testCase.expectedPath
					if !relative {
						expectedURL = tr.base + expectedURL
					}
				} else {
					urlBase, err := url.Parse(tr.base)
					if err != nil {
						t.Fatal(err)
					}
					urlBase.Scheme = proto[0]
					expectedURL = testCase.expectedPath
					if !relative {
						expectedURL = urlBase.String() + expectedURL
					}
				}

				if buildURL != expectedURL {
					t.Errorf("[relative=%t, request=%q, case=%q]: %q != %q", relative, tr.name, testCase.description, buildURL, expectedURL)
				}
			}
		}
	}

	doTest(true)
	doTest(false)
}

func TestBuilderFromRequestWithPrefix(t *testing.T) {
	u, err := url.Parse("http://example.com/prefix/v2/")
	if err != nil {
		t.Fatal(err)
	}

	forwardedProtoHeader := make(http.Header, 1)
	forwardedProtoHeader.Set("X-Forwarded-Proto", "https")

	testRequests := []struct {
		request    *http.Request
		base       string
		configHost url.URL
	}{
		{
			request: &http.Request{URL: u, Host: u.Host},
			base:    "http://example.com/prefix/",
		},

		{
			request: &http.Request{URL: u, Host: u.Host, Header: forwardedProtoHeader},
			base:    "http://example.com/prefix/",
		},
		{
			request: &http.Request{URL: u, Host: u.Host, Header: forwardedProtoHeader},
			base:    "https://example.com/prefix/",
		},
		{
			request: &http.Request{URL: u, Host: u.Host, Header: forwardedProtoHeader},
			base:    "https://subdomain.example.com/prefix/",
			configHost: url.URL{
				Scheme: "https",
				Host:   "subdomain.example.com",
				Path:   "/prefix/",
			},
		},
	}

	var relative bool
	for _, tr := range testRequests {
		var builder *URLBuilder
		if tr.configHost.Scheme != "" && tr.configHost.Host != "" {
			builder = NewURLBuilder(&tr.configHost, false)
		} else {
			builder = NewURLBuilderFromRequest(tr.request, false)
		}

		for _, testCase := range makeURLBuilderTestCases(builder) {
			buildURL, err := testCase.build()
			if err != nil {
				t.Fatalf("%s: error building url: %v", testCase.description, err)
			}

			var expectedURL string
			proto, ok := tr.request.Header["X-Forwarded-Proto"]
			if !ok {
				expectedURL = testCase.expectedPath
				if !relative {
					expectedURL = tr.base[0:len(tr.base)-1] + expectedURL
				}
			} else {
				urlBase, err := url.Parse(tr.base)
				if err != nil {
					t.Fatal(err)
				}
				urlBase.Scheme = proto[0]
				expectedURL = testCase.expectedPath
				if !relative {
					expectedURL = urlBase.String()[0:len(urlBase.String())-1] + expectedURL
				}

			}

			if buildURL != expectedURL {
				t.Fatalf("%s: %q != %q", testCase.description, buildURL, expectedURL)
			}
		}
	}
}

func TestIsIPv6Address(t *testing.T) {
	for _, tc := range []struct {
		name    string
		address string
		isIPv6  bool
	}{
		{
			name:    "IPv6 short address",
			address: `2607:f0d0:1002:51::4`,
			isIPv6:  true,
		},
		{
			name:    "IPv6 short address enclosed in brackets",
			address: "[2607:f0d0:1002:51::4]",
			isIPv6:  true,
		},
		{
			name:    "IPv6 address",
			address: `2607:f0d0:1002:0051:0000:0000:0000:0004`,
			isIPv6:  true,
		},
		{
			name:    "IPv6 address with numeric zone ID",
			address: `fe80::bd0f:a8bc:6480:238b%11`,
			isIPv6:  true,
		},
		{
			name:    "IPv6 address with device name as zone ID",
			address: `fe80::bd0f:a8bc:6480:238b%eth0`,
			isIPv6:  true,
		},
		{
			name:    "IPv6 address with device name as zone ID enclosed in brackets",
			address: `[fe80::bd0f:a8bc:6480:238b%eth0]`,
			isIPv6:  true,
		},
		{
			name:    "IPv4-mapped address",
			address: "::FFFF:129.144.52.38",
			isIPv6:  true,
		},
		{
			name:    "localhost",
			address: "::1",
			isIPv6:  true,
		},
		{
			name:    "localhost",
			address: "::1",
			isIPv6:  true,
		},
		{
			name:    "long localhost address",
			address: "0:0:0:0:0:0:0:1",
			isIPv6:  true,
		},
		{
			name:    "IPv6 long address with port",
			address: "[2607:f0d0:1002:0051:0000:0000:0000:0004]:4321",
			isIPv6:  false,
		},
		{
			name:    "too many groups",
			address: "2607:f0d0:1002:0051:0000:0000:0000:0004:4321",
			isIPv6:  false,
		},
		{
			name:    "square brackets don't make an IPv6 address",
			address: "[2607:f0d0]",
			isIPv6:  false,
		},
		{
			name:    "require two consecutive colons in localhost",
			address: ":1",
			isIPv6:  false,
		},
		{
			name:    "more then 4 hexadecimal digits",
			address: "2607:f0d0b:1002:0051:0000:0000:0000:0004",
			isIPv6:  false,
		},
		{
			name:    "too short address",
			address: `2607:f0d0:1002:0000:0000:0000:0004`,
			isIPv6:  false,
		},
		{
			name:    "IPv4 address",
			address: `192.168.100.1`,
			isIPv6:  false,
		},
		{
			name:    "unclosed bracket",
			address: `[2607:f0d0:1002:0051:0000:0000:0000:0004`,
			isIPv6:  false,
		},
		{
			name:    "trailing bracket",
			address: `2607:f0d0:1002:0051:0000:0000:0000:0004]`,
			isIPv6:  false,
		},
		{
			name:    "domain name",
			address: `localhost`,
			isIPv6:  false,
		},
	} {
		isIPv6 := isIPv6Address(tc.address)
		if isIPv6 && !tc.isIPv6 {
			t.Errorf("[%s] address %q falsely detected as IPv6 address", tc.name, tc.address)
		} else if !isIPv6 && tc.isIPv6 {
			t.Errorf("[%s] address %q not recognized as IPv6", tc.name, tc.address)
		}
	}
}
