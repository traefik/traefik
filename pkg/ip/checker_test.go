package ip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAuthorized(t *testing.T) {
	testCases := []struct {
		desc       string
		whiteList  []string
		remoteAddr string
		authorized bool
	}{
		{
			desc:       "remoteAddr not in range",
			whiteList:  []string{"1.2.3.4/24"},
			remoteAddr: "10.2.3.1:123",
			authorized: false,
		},
		{
			desc:       "remoteAddr in range",
			whiteList:  []string{"1.2.3.4/24"},
			remoteAddr: "1.2.3.1:123",
			authorized: true,
		},
		{
			desc:       "octal ip in remoteAddr",
			whiteList:  []string{"127.2.3.4/24"},
			remoteAddr: "0127.2.3.1:123",
			authorized: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ipChecker, err := NewChecker(test.whiteList)
			require.NoError(t, err)

			err = ipChecker.IsAuthorized(test.remoteAddr)
			if test.authorized {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	testCases := []struct {
		desc                  string
		trustedIPs            []string
		expectedAuthorizedIPs []*net.IPNet
		errMessage            string
	}{
		{
			desc:                  "nil trusted IPs",
			trustedIPs:            nil,
			expectedAuthorizedIPs: nil,
			errMessage:            "no trusted IPs provided",
		}, {
			desc:                  "empty trusted IPs",
			trustedIPs:            []string{},
			expectedAuthorizedIPs: nil,
			errMessage:            "no trusted IPs provided",
		}, {
			desc: "trusted IPs containing empty string",
			trustedIPs: []string{
				"1.2.3.4/24",
				"",
				"fe80::/16",
			},
			expectedAuthorizedIPs: nil,
			errMessage:            "parsing CIDR trusted IPs <nil>: invalid CIDR address: ",
		}, {
			desc: "trusted IPs containing only an empty string",
			trustedIPs: []string{
				"",
			},
			expectedAuthorizedIPs: nil,
			errMessage:            "parsing CIDR trusted IPs <nil>: invalid CIDR address: ",
		}, {
			desc: "trusted IPs containing an invalid string",
			trustedIPs: []string{
				"foo",
			},
			expectedAuthorizedIPs: nil,
			errMessage:            "parsing CIDR trusted IPs <nil>: invalid CIDR address: foo",
		}, {
			desc: "IPv4 & IPv6 trusted IPs",
			trustedIPs: []string{
				"1.2.3.4/24",
				"fe80::/16",
			},
			expectedAuthorizedIPs: []*net.IPNet{
				{IP: net.IPv4(1, 2, 3, 0).To4(), Mask: net.IPv4Mask(255, 255, 255, 0)},
				{IP: net.ParseIP("fe80::"), Mask: net.IPMask(net.ParseIP("ffff::"))},
			},
			errMessage: "",
		}, {
			desc: "IPv4 only",
			trustedIPs: []string{
				"127.0.0.1/8",
			},
			expectedAuthorizedIPs: []*net.IPNet{
				{IP: net.IPv4(127, 0, 0, 0).To4(), Mask: net.IPv4Mask(255, 0, 0, 0)},
			},
			errMessage: "",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ipChecker, err := NewChecker(test.trustedIPs)
			if test.errMessage != "" {
				require.EqualError(t, err, test.errMessage)
			} else {
				require.NoError(t, err)
				for index, actual := range ipChecker.authorizedIPsNet {
					expected := test.expectedAuthorizedIPs[index]
					assert.Equal(t, expected.IP, actual.IP)
					assert.Equal(t, expected.Mask.String(), actual.Mask.String())
				}
			}
		})
	}
}

func TestContainsIsAllowed(t *testing.T) {
	testCases := []struct {
		desc       string
		trustedIPs []string
		passIPs    []string
		rejectIPs  []string
	}{
		{
			desc:       "IPv4",
			trustedIPs: []string{"1.2.3.4/24"},
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
			desc:       "IPv4 single IP",
			trustedIPs: []string{"8.8.8.8"},
			passIPs:    []string{"8.8.8.8"},
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
			desc:       "IPv4 Net single IP",
			trustedIPs: []string{"8.8.8.8/32"},
			passIPs:    []string{"8.8.8.8"},
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
			desc:       "multiple IPv4",
			trustedIPs: []string{"1.2.3.4/24", "8.8.8.8/8"},
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
			desc:       "IPv6",
			trustedIPs: []string{"2a03:4000:6:d080::/64"},
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
			desc:       "IPv6 single IP",
			trustedIPs: []string{"2a03:4000:6:d080::42/128"},
			passIPs:    []string{"2a03:4000:6:d080::42"},
			rejectIPs: []string{
				"2a03:4000:6:d080::1",
				"2a03:4000:6:d080:dead:beef:ffff:ffff",
				"2a03:4000:6:d080::43",
			},
		},
		{
			desc:       "multiple IPv6",
			trustedIPs: []string{"2a03:4000:6:d080::/64", "fe80::/16"},
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
			desc:       "multiple IPv6 & IPv4",
			trustedIPs: []string{"2a03:4000:6:d080::/64", "fe80::/16", "1.2.3.4/24", "8.8.8.8/8"},
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
			desc:       "broken IP-addresses",
			trustedIPs: []string{"127.0.0.1/32"},
			passIPs:    nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ipChecker, err := NewChecker(test.trustedIPs)

			require.NoError(t, err)
			require.NotNil(t, ipChecker)

			for _, testIP := range test.passIPs {
				allowed, err := ipChecker.Contains(testIP)
				require.NoError(t, err)
				assert.Truef(t, allowed, "%s should have passed.", testIP)
			}

			for _, testIP := range test.rejectIPs {
				allowed, err := ipChecker.Contains(testIP)
				require.NoError(t, err)
				assert.Falsef(t, allowed, "%s should not have passed.", testIP)
			}
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

	ipChecker, err := NewChecker([]string{"1.2.3.4/24"})
	require.NoError(t, err)

	for _, testIP := range brokenIPs {
		_, err := ipChecker.Contains(testIP)
		assert.Error(t, err)
	}
}
