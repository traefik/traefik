package whitelist

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAuthorized(t *testing.T) {
	testCases := []struct {
		desc                string
		whiteList           []string
		allowXForwardedFor  bool
		remoteAddr          string
		xForwardedForValues []string
		authorized          bool
	}{
		{
			desc:                "allow UseXForwardedFor, remoteAddr not in range, UseXForwardedFor in range",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  true,
			remoteAddr:          "10.2.3.1:123",
			xForwardedForValues: []string{"1.2.3.1", "10.2.3.1"},
			authorized:          true,
		},
		{
			desc:                "allow UseXForwardedFor, remoteAddr not in range, UseXForwardedFor in range (compact XFF)",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  true,
			remoteAddr:          "10.2.3.1:123",
			xForwardedForValues: []string{"10.2.3.1, 1.2.3.1"},
			authorized:          true,
		},
		{
			desc:                "allow UseXForwardedFor, remoteAddr in range, UseXForwardedFor in range",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  true,
			remoteAddr:          "1.2.3.1:123",
			xForwardedForValues: []string{"1.2.3.1", "10.2.3.1"},
			authorized:          true,
		},
		{
			desc:                "allow UseXForwardedFor, remoteAddr in range, UseXForwardedFor not in range",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  true,
			remoteAddr:          "1.2.3.1:123",
			xForwardedForValues: []string{"10.2.3.1", "10.2.3.1"},
			authorized:          true,
		},
		{
			desc:                "allow UseXForwardedFor, remoteAddr not in range, UseXForwardedFor not in range",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  true,
			remoteAddr:          "10.2.3.1:123",
			xForwardedForValues: []string{"10.2.3.1", "10.2.3.1"},
			authorized:          false,
		},
		{
			desc:                "don't allow UseXForwardedFor, remoteAddr not in range, UseXForwardedFor in range",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  false,
			remoteAddr:          "10.2.3.1:123",
			xForwardedForValues: []string{"1.2.3.1", "10.2.3.1"},
			authorized:          false,
		},
		{
			desc:                "don't allow UseXForwardedFor, remoteAddr in range, UseXForwardedFor in range",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  false,
			remoteAddr:          "1.2.3.1:123",
			xForwardedForValues: []string{"1.2.3.1", "10.2.3.1"},
			authorized:          true,
		},
		{
			desc:                "don't allow UseXForwardedFor, remoteAddr in range, UseXForwardedFor not in range",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  false,
			remoteAddr:          "1.2.3.1:123",
			xForwardedForValues: []string{"10.2.3.1", "10.2.3.1"},
			authorized:          true,
		},
		{
			desc:                "don't allow UseXForwardedFor, remoteAddr not in range, UseXForwardedFor not in range",
			whiteList:           []string{"1.2.3.4/24"},
			allowXForwardedFor:  false,
			remoteAddr:          "10.2.3.1:123",
			xForwardedForValues: []string{"10.2.3.1", "10.2.3.1"},
			authorized:          false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := NewRequest(test.remoteAddr, test.xForwardedForValues)

			whiteLister, err := NewIP(test.whiteList, false, test.allowXForwardedFor)
			require.NoError(t, err)

			err = whiteLister.IsAuthorized(req)
			if test.authorized {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	cases := []struct {
		desc               string
		whiteList          []string
		expectedWhitelists []*net.IPNet
		errMessage         string
	}{
		{
			desc:               "nil whitelist",
			whiteList:          nil,
			expectedWhitelists: nil,
			errMessage:         "no white list provided",
		}, {
			desc:               "empty whitelist",
			whiteList:          []string{},
			expectedWhitelists: nil,
			errMessage:         "no white list provided",
		}, {
			desc: "whitelist containing empty string",
			whiteList: []string{
				"1.2.3.4/24",
				"",
				"fe80::/16",
			},
			expectedWhitelists: nil,
			errMessage:         "parsing CIDR white list <nil>: invalid CIDR address: ",
		}, {
			desc: "whitelist containing only an empty string",
			whiteList: []string{
				"",
			},
			expectedWhitelists: nil,
			errMessage:         "parsing CIDR white list <nil>: invalid CIDR address: ",
		}, {
			desc: "whitelist containing an invalid string",
			whiteList: []string{
				"foo",
			},
			expectedWhitelists: nil,
			errMessage:         "parsing CIDR white list <nil>: invalid CIDR address: foo",
		}, {
			desc: "IPv4 & IPv6 whitelist",
			whiteList: []string{
				"1.2.3.4/24",
				"fe80::/16",
			},
			expectedWhitelists: []*net.IPNet{
				{IP: net.IPv4(1, 2, 3, 0).To4(), Mask: net.IPv4Mask(255, 255, 255, 0)},
				{IP: net.ParseIP("fe80::"), Mask: net.IPMask(net.ParseIP("ffff::"))},
			},
			errMessage: "",
		}, {
			desc: "IPv4 only",
			whiteList: []string{
				"127.0.0.1/8",
			},
			expectedWhitelists: []*net.IPNet{
				{IP: net.IPv4(127, 0, 0, 0).To4(), Mask: net.IPv4Mask(255, 0, 0, 0)},
			},
			errMessage: "",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			whiteLister, err := NewIP(test.whiteList, false, false)
			if test.errMessage != "" {
				require.EqualError(t, err, test.errMessage)
			} else {
				require.NoError(t, err)
				for index, actual := range whiteLister.whiteListsNet {
					expected := test.expectedWhitelists[index]
					assert.Equal(t, expected.IP, actual.IP)
					assert.Equal(t, expected.Mask.String(), actual.Mask.String())
				}
			}
		})
	}
}

func TestContainsIsAllowed(t *testing.T) {
	cases := []struct {
		desc             string
		whitelistStrings []string
		passIPs          []string
		rejectIPs        []string
	}{
		{
			desc:             "IPv4",
			whitelistStrings: []string{"1.2.3.4/24"},
			passIPs: []string{
				"1.2.3.1",
				"1.2.3.32",
				"1.2.3.156",
				"1.2.3.255",
			},
			rejectIPs: []string{
				"1.2.16.1",
				"1.2.32.1",
				"127.0.0.1",
				"8.8.8.8",
			},
		},
		{
			desc:             "IPv4 single IP",
			whitelistStrings: []string{"8.8.8.8"},
			passIPs:          []string{"8.8.8.8"},
			rejectIPs: []string{
				"8.8.8.7",
				"8.8.8.9",
				"8.8.8.0",
				"8.8.8.255",
				"4.4.4.4",
				"127.0.0.1",
			},
		},
		{
			desc:             "IPv4 Net single IP",
			whitelistStrings: []string{"8.8.8.8/32"},
			passIPs:          []string{"8.8.8.8"},
			rejectIPs: []string{
				"8.8.8.7",
				"8.8.8.9",
				"8.8.8.0",
				"8.8.8.255",
				"4.4.4.4",
				"127.0.0.1",
			},
		},
		{
			desc:             "multiple IPv4",
			whitelistStrings: []string{"1.2.3.4/24", "8.8.8.8/8"},
			passIPs: []string{
				"1.2.3.1",
				"1.2.3.32",
				"1.2.3.156",
				"1.2.3.255",
				"8.8.4.4",
				"8.0.0.1",
				"8.32.42.128",
				"8.255.255.255",
			},
			rejectIPs: []string{
				"1.2.16.1",
				"1.2.32.1",
				"127.0.0.1",
				"4.4.4.4",
				"4.8.8.8",
			},
		},
		{
			desc:             "IPv6",
			whitelistStrings: []string{"2a03:4000:6:d080::/64"},
			passIPs: []string{
				"2a03:4000:6:d080::",
				"2a03:4000:6:d080::1",
				"2a03:4000:6:d080:dead:beef:ffff:ffff",
				"2a03:4000:6:d080::42",
			},
			rejectIPs: []string{
				"2a03:4000:7:d080::",
				"2a03:4000:7:d080::1",
				"fe80::",
				"4242::1",
			},
		},
		{
			desc:             "IPv6 single IP",
			whitelistStrings: []string{"2a03:4000:6:d080::42/128"},
			passIPs:          []string{"2a03:4000:6:d080::42"},
			rejectIPs: []string{
				"2a03:4000:6:d080::1",
				"2a03:4000:6:d080:dead:beef:ffff:ffff",
				"2a03:4000:6:d080::43",
			},
		},
		{
			desc:             "multiple IPv6",
			whitelistStrings: []string{"2a03:4000:6:d080::/64", "fe80::/16"},
			passIPs: []string{
				"2a03:4000:6:d080::",
				"2a03:4000:6:d080::1",
				"2a03:4000:6:d080:dead:beef:ffff:ffff",
				"2a03:4000:6:d080::42",
				"fe80::1",
				"fe80:aa00:00bb:4232:ff00:eeee:00ff:1111",
				"fe80::fe80",
			},
			rejectIPs: []string{
				"2a03:4000:7:d080::",
				"2a03:4000:7:d080::1",
				"4242::1",
			},
		},
		{
			desc:             "multiple IPv6 & IPv4",
			whitelistStrings: []string{"2a03:4000:6:d080::/64", "fe80::/16", "1.2.3.4/24", "8.8.8.8/8"},
			passIPs: []string{
				"2a03:4000:6:d080::",
				"2a03:4000:6:d080::1",
				"2a03:4000:6:d080:dead:beef:ffff:ffff",
				"2a03:4000:6:d080::42",
				"fe80::1",
				"fe80:aa00:00bb:4232:ff00:eeee:00ff:1111",
				"fe80::fe80",
				"1.2.3.1",
				"1.2.3.32",
				"1.2.3.156",
				"1.2.3.255",
				"8.8.4.4",
				"8.0.0.1",
				"8.32.42.128",
				"8.255.255.255",
			},
			rejectIPs: []string{
				"2a03:4000:7:d080::",
				"2a03:4000:7:d080::1",
				"4242::1",
				"1.2.16.1",
				"1.2.32.1",
				"127.0.0.1",
				"4.4.4.4",
				"4.8.8.8",
			},
		},
		{
			desc:             "broken IP-addresses",
			whitelistStrings: []string{"127.0.0.1/32"},
			passIPs:          nil,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			whiteLister, err := NewIP(test.whitelistStrings, false, false)

			require.NoError(t, err)
			require.NotNil(t, whiteLister)

			for _, testIP := range test.passIPs {
				allowed, err := whiteLister.contains(testIP)
				require.NoError(t, err)
				assert.Truef(t, allowed, "%s should have passed.", testIP)
			}

			for _, testIP := range test.rejectIPs {
				allowed, err := whiteLister.contains(testIP)
				require.NoError(t, err)
				assert.Falsef(t, allowed, "%s should not have passed.", testIP)
			}
		})
	}
}

func TestContainsInsecure(t *testing.T) {
	mustNewIP := func(whitelistStrings []string, insecure bool) *IP {
		ip, err := NewIP(whitelistStrings, insecure, false)
		if err != nil {
			t.Fatal(err)
		}
		return ip
	}

	testCases := []struct {
		desc        string
		whiteLister *IP
		ip          string
		expected    bool
	}{
		{
			desc:        "valid ip and insecure",
			whiteLister: mustNewIP([]string{"1.2.3.4/24"}, true),
			ip:          "1.2.3.1",
			expected:    true,
		},
		{
			desc:        "invalid ip and insecure",
			whiteLister: mustNewIP([]string{"1.2.3.4/24"}, true),
			ip:          "10.2.3.1",
			expected:    true,
		},
		{
			desc:        "invalid ip and secure",
			whiteLister: mustNewIP([]string{"1.2.3.4/24"}, false),
			ip:          "10.2.3.1",
			expected:    false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ok, err := test.whiteLister.contains(test.ip)
			require.NoError(t, err)

			assert.Equal(t, test.expected, ok)
		})
	}
}

func TestContainsBrokenIPs(t *testing.T) {
	brokenIPs := []string{
		"foo",
		"10.0.0.350",
		"fe:::80",
		"",
		"\\&$§&/(",
	}

	whiteLister, err := NewIP([]string{"1.2.3.4/24"}, false, false)
	require.NoError(t, err)

	for _, testIP := range brokenIPs {
		_, err := whiteLister.contains(testIP)
		assert.Error(t, err)
	}
}

func NewRequest(remoteAddr string, xForwardedFor []string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	if len(remoteAddr) > 0 {
		req.RemoteAddr = remoteAddr
	}
	if len(xForwardedFor) > 0 {
		for _, xff := range xForwardedFor {
			req.Header.Add(XForwardedFor, xff)
		}
	}
	return req
}
