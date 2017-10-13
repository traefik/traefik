package whitelist

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cases := []struct {
		desc               string
		whitelistStrings   []string
		expectedWhitelists []*net.IPNet
		errMessage         string
	}{
		{
			desc:               "nil whitelist",
			whitelistStrings:   nil,
			expectedWhitelists: nil,
			errMessage:         "no whiteListsNet provided",
		}, {
			desc:               "empty whitelist",
			whitelistStrings:   []string{},
			expectedWhitelists: nil,
			errMessage:         "no whiteListsNet provided",
		}, {
			desc: "whitelist containing empty string",
			whitelistStrings: []string{
				"1.2.3.4/24",
				"",
				"fe80::/16",
			},
			expectedWhitelists: nil,
			errMessage:         "parsing CIDR whitelist <nil>: invalid CIDR address: ",
		}, {
			desc: "whitelist containing only an empty string",
			whitelistStrings: []string{
				"",
			},
			expectedWhitelists: nil,
			errMessage:         "parsing CIDR whitelist <nil>: invalid CIDR address: ",
		}, {
			desc: "whitelist containing an invalid string",
			whitelistStrings: []string{
				"foo",
			},
			expectedWhitelists: nil,
			errMessage:         "parsing CIDR whitelist <nil>: invalid CIDR address: foo",
		}, {
			desc: "IPv4 & IPv6 whitelist",
			whitelistStrings: []string{
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
			whitelistStrings: []string{
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
			whitelister, err := NewIP(test.whitelistStrings)
			if test.errMessage != "" {
				require.EqualError(t, err, test.errMessage)
			} else {
				require.NoError(t, err)
				for index, actual := range whitelister.whiteListsNet {
					expected := test.expectedWhitelists[index]
					assert.Equal(t, expected.IP, actual.IP)
					assert.Equal(t, expected.Mask.String(), actual.Mask.String())
				}
			}
		})
	}
}

func TestIsAllowed(t *testing.T) {
	cases := []struct {
		desc             string
		whitelistStrings []string
		passIPs          []string
		rejectIPs        []string
	}{
		{
			desc: "IPv4",
			whitelistStrings: []string{
				"1.2.3.4/24",
			},
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
			desc: "IPv4 single IP",
			whitelistStrings: []string{
				"8.8.8.8",
			},
			passIPs: []string{
				"8.8.8.8",
			},
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
			desc: "IPv4 Net single IP",
			whitelistStrings: []string{
				"8.8.8.8/32",
			},
			passIPs: []string{
				"8.8.8.8",
			},
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
			desc: "multiple IPv4",
			whitelistStrings: []string{
				"1.2.3.4/24",
				"8.8.8.8/8",
			},
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
			desc: "IPv6",
			whitelistStrings: []string{
				"2a03:4000:6:d080::/64",
			},
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
			desc: "IPv6 single IP",
			whitelistStrings: []string{
				"2a03:4000:6:d080::42/128",
			},
			passIPs: []string{
				"2a03:4000:6:d080::42",
			},
			rejectIPs: []string{
				"2a03:4000:6:d080::1",
				"2a03:4000:6:d080:dead:beef:ffff:ffff",
				"2a03:4000:6:d080::43",
			},
		},
		{
			desc: "multiple IPv6",
			whitelistStrings: []string{
				"2a03:4000:6:d080::/64",
				"fe80::/16",
			},
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
			desc: "multiple IPv6 & IPv4",
			whitelistStrings: []string{
				"2a03:4000:6:d080::/64",
				"fe80::/16",
				"1.2.3.4/24",
				"8.8.8.8/8",
			},
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
			desc: "broken IP-addresses",
			whitelistStrings: []string{
				"127.0.0.1/32",
			},
			passIPs: nil,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			whiteLister, err := NewIP(test.whitelistStrings)

			require.NoError(t, err)
			require.NotNil(t, whiteLister)

			for _, testIP := range test.passIPs {
				allowed, ip, err := whiteLister.Contains(testIP)
				require.NoError(t, err)
				require.NotNil(t, ip, err)
				assert.True(t, allowed, testIP+" should have passed "+test.desc)
			}

			for _, testIP := range test.rejectIPs {
				allowed, ip, err := whiteLister.Contains(testIP)
				require.NoError(t, err)
				require.NotNil(t, ip, err)
				assert.False(t, allowed, testIP+" should not have passed "+test.desc)
			}
		})
	}
}

func TestBrokenIPs(t *testing.T) {
	brokenIPs := []string{
		"foo",
		"10.0.0.350",
		"fe:::80",
		"",
		"\\&$§&/(",
	}

	whiteLister, err := NewIP([]string{"1.2.3.4/24"})
	require.NoError(t, err)

	for _, testIP := range brokenIPs {
		_, ip, err := whiteLister.Contains(testIP)
		assert.Error(t, err)
		require.Nil(t, ip, err)
	}

}
