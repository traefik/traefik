package ip

import (
	"fmt"
	"net"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAuthorized(t *testing.T) {
	testCases := []struct {
		desc       string
		allowList  []string
		remoteAddr string
		authorized bool
	}{
		{
			desc:       "remoteAddr not in range",
			allowList:  []string{"1.2.3.4/24"},
			remoteAddr: "10.2.3.1:123",
			authorized: false,
		},
		{
			desc:       "remoteAddr in range",
			allowList:  []string{"1.2.3.4/24"},
			remoteAddr: "1.2.3.1:123",
			authorized: true,
		},
		{
			desc:       "octal ip in remoteAddr",
			allowList:  []string{"127.2.3.4/24"},
			remoteAddr: "0127.2.3.1:123",
			authorized: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ipChecker, err := NewChecker(test.allowList)
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
			errMessage:            `parsing CIDR trusted IPs : netip.ParsePrefix(""): no '/'`,
		}, {
			desc: "trusted IPs containing only an empty string",
			trustedIPs: []string{
				"",
			},
			expectedAuthorizedIPs: nil,
			errMessage:            `parsing CIDR trusted IPs : netip.ParsePrefix(""): no '/'`,
		}, {
			desc: "trusted IPs containing an invalid string",
			trustedIPs: []string{
				"foo",
			},
			expectedAuthorizedIPs: nil,
			errMessage:            `parsing CIDR trusted IPs foo: netip.ParsePrefix("foo"): no '/'`,
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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ipChecker, err := NewChecker(test.trustedIPs)
			if test.errMessage != "" {
				require.EqualError(t, err, test.errMessage)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ipChecker)
				for _, expected := range test.expectedAuthorizedIPs {
					assert.True(t, ipChecker.ContainsIP(expected.IP), "expected %s to be authorized", expected.IP)
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
				"fe80:aa00:00bb:4232:ff00:eeee:00ff:1111%vEthernet",
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
				"2a03:4000:7:d080::1%vmnet1",
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

func numToIP4String(n uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

func generateNonAdjacentCIDRs(n uint32) (cidrs []string, matchingIP string, nonMatchingIP string, err error) {
	var maxNumOfAddresses uint32 = 1<<32 - 1
	var gap uint32 = 512 // gap of 512 (2 /24 blocks) ensures non-overlapping and non-adjacent CIDRs
	maxNumOfCIDRs := maxNumOfAddresses / gap
	if n <= 0 || n > maxNumOfCIDRs {
		return nil, "", "", fmt.Errorf("n must be between 1 and %d", maxNumOfCIDRs)
	}
	cidrs = make([]string, n)
	ip := uint32(0)
	for i := range n {
		cidrs[i] = numToIP4String(ip) + "/24"
		ip += gap
	}
	mid := (n/2)&^1*gap + gap/4
	matchingIP = numToIP4String(mid)
	nonMatchingIP = numToIP4String(gap / 4 * 3)
	return
}

func generateUniqueSingleIPs(n uint32) (ips []string, matchingIP string, nonMatchingIP string, err error) {
	var maxNumOfAddresses uint32 = 1<<32 - 1
	if n <= 0 || n > maxNumOfAddresses {
		return nil, "", "", fmt.Errorf("n must be between 1 and %d", maxNumOfAddresses)
	}
	ips = make([]string, n)
	for i := range n {
		ips[i] = numToIP4String(i)
	}
	matchingIP = numToIP4String(n / 2)
	nonMatchingIP = numToIP4String(n)
	return
}

func newTestChecker(b *testing.B, trustedIPs []string) *Checker {
	b.Helper()
	checker, err := NewChecker(trustedIPs)
	require.NoError(b, err)
	return checker
}

func mustParseIP(b *testing.B, s string) net.IP {
	b.Helper()
	ip := net.ParseIP(s)
	require.NotNil(b, ip, "failed to parse IP: %s", s)
	return ip
}

func requireContainsIP(b *testing.B, checker *Checker, ip net.IP) {
	b.Helper()
	require.True(b, checker.ContainsIP(ip), "test IP should be in the trusted list")
}

func requireNotContainsIP(b *testing.B, checker *Checker, ip net.IP) {
	b.Helper()
	require.False(b, checker.ContainsIP(ip), "test IP should not be in the trusted list")
}

func BenchmarkContainsIP_MatchInSingleIPList_100k(b *testing.B) {
	trustedIPs, matchingIP, _, err := generateUniqueSingleIPs(100000)
	require.NoError(b, err)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, matchingIP)
	requireContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func BenchmarkContainsIP_NoMatchInSingleIPList_100k(b *testing.B) {
	trustedIPs, _, nonMatchingIP, err := generateUniqueSingleIPs(100000)
	require.NoError(b, err)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, nonMatchingIP)
	requireNotContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func BenchmarkContainsIP_MatchInCIDRList_10k(b *testing.B) {
	trustedIPs, matchingIP, _, err := generateNonAdjacentCIDRs(10000)
	require.NoError(b, err)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, matchingIP)
	requireContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func BenchmarkContainsIP_NoMatchInCIDRList_10k(b *testing.B) {
	trustedIPs, _, nonMatchingIP, err := generateNonAdjacentCIDRs(10000)
	require.NoError(b, err)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, nonMatchingIP)
	requireNotContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func BenchmarkContainsIP_MatchInCIDRList_100k(b *testing.B) {
	trustedIPs, matchingIP, _, err := generateNonAdjacentCIDRs(100000)
	require.NoError(b, err)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, matchingIP)
	requireContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func BenchmarkContainsIP_NoMatchInCIDRList_100k(b *testing.B) {
	trustedIPs, _, nonMatchingIP, err := generateNonAdjacentCIDRs(100000)
	require.NoError(b, err)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, nonMatchingIP)
	requireNotContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func BenchmarkContainsIP_MatchInMixedList_SingleIP_100k(b *testing.B) {
	cidrs, _, _, err := generateNonAdjacentCIDRs(50000)
	require.NoError(b, err)
	singles, matchingSingleIP, _, err := generateUniqueSingleIPs(50000)
	require.NoError(b, err)

	trustedIPs := append(cidrs, singles...)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, matchingSingleIP)
	requireContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func BenchmarkContainsIP_MatchInMixedList_CIDR_100k(b *testing.B) {
	cidrs, matchingCIDRIP, _, err := generateNonAdjacentCIDRs(50000)
	require.NoError(b, err)
	singles, _, _, err := generateUniqueSingleIPs(50000)
	require.NoError(b, err)

	trustedIPs := append(cidrs, singles...)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, matchingCIDRIP)
	requireContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func BenchmarkContainsIP_NoMatchInMixedList_100k(b *testing.B) {
	cidrs, _, _, err := generateNonAdjacentCIDRs(50000)
	require.NoError(b, err)
	singles, _, _, err := generateUniqueSingleIPs(50000)
	require.NoError(b, err)

	trustedIPs := append(cidrs, singles...)

	checker := newTestChecker(b, trustedIPs)
	testIP := mustParseIP(b, numToIP4String(1<<32-1))
	requireNotContainsIP(b, checker, testIP)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.ContainsIP(testIP)
	}
}

func TestNetworkLastAddress(t *testing.T) {
	testCases := []struct {
		desc     string
		prefix   netip.Prefix
		expected netip.Addr
	}{
		{
			desc:     "IPv4 /24",
			prefix:   netip.MustParsePrefix("192.168.1.0/24"),
			expected: netip.MustParseAddr("192.168.1.255"),
		},
		{
			desc:     "IPv4 /0 (all addresses)",
			prefix:   netip.MustParsePrefix("0.0.0.0/0"),
			expected: netip.MustParseAddr("255.255.255.255"),
		},
		{
			desc:     "IPv4 /8",
			prefix:   netip.MustParsePrefix("10.0.0.0/8"),
			expected: netip.MustParseAddr("10.255.255.255"),
		},
		{
			desc:     "IPv4 /32 single IP",
			prefix:   netip.MustParsePrefix("10.0.0.1/32"),
			expected: netip.MustParseAddr("10.0.0.1"),
		},
		{
			desc:     "IPv4 /16",
			prefix:   netip.MustParsePrefix("172.16.0.0/16"),
			expected: netip.MustParseAddr("172.16.255.255"),
		},
		{
			desc:     "IPv4 /30",
			prefix:   netip.MustParsePrefix("192.168.0.0/30"),
			expected: netip.MustParseAddr("192.168.0.3"),
		},
		{
			desc:     "IPv4 /29",
			prefix:   netip.MustParsePrefix("10.10.10.16/29"),
			expected: netip.MustParseAddr("10.10.10.23"),
		},
		{
			desc:     "IPv4 /28 non-aligned",
			prefix:   netip.MustParsePrefix("10.10.10.20/28"),
			expected: netip.MustParseAddr("10.10.10.31"),
		},
		{
			desc:     "IPv6 /64",
			prefix:   netip.MustParsePrefix("2001:db8::/64"),
			expected: netip.MustParseAddr("2001:db8::ffff:ffff:ffff:ffff"),
		},
		{
			desc:     "IPv6 /16",
			prefix:   netip.MustParsePrefix("fe80::/16"),
			expected: netip.MustParseAddr("fe80:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		},
		{
			desc:     "IPv6 /128 single IP",
			prefix:   netip.MustParsePrefix("2001:db8::1/128"),
			expected: netip.MustParseAddr("2001:db8::1"),
		},
		{
			desc:     "IPv6 /0 (all addresses)",
			prefix:   netip.MustParsePrefix("::/0"),
			expected: netip.MustParseAddr("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		},
		{
			desc:     "IPv6 /48",
			prefix:   netip.MustParsePrefix("2a03:4000:6:d080::/48"),
			expected: netip.MustParseAddr("2a03:4000:6:ffff:ffff:ffff:ffff:ffff"),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := networkLastAddress(test.prefix)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestMergeRanges(t *testing.T) {
	testCases := []struct {
		desc           string
		trustedIPs     []string
		expectedRanges int
		passIPs        []string
		rejectIPs      []string
	}{
		{
			desc: "overlapping ranges merge",
			trustedIPs: []string{
				"10.0.0.0/24",
				"10.0.0.0/16",
			},
			expectedRanges: 1,
			passIPs: []string{
				"10.0.0.1",
				"10.0.255.255",
			},
			rejectIPs: []string{
				"10.1.0.1",
			},
		},
		{
			desc: "adjacent ranges merge",
			trustedIPs: []string{
				"10.0.0.0/25",
				"10.0.0.128/25",
			},
			expectedRanges: 1,
			passIPs: []string{
				"10.0.0.1",
				"10.0.0.128",
				"10.0.0.255",
			},
			rejectIPs: []string{
				"10.0.1.1",
			},
		},
		{
			desc: "subset ranges merge",
			trustedIPs: []string{
				"10.0.0.0/8",
				"10.0.0.0/24",
				"10.0.1.0/24",
			},
			expectedRanges: 1,
			passIPs: []string{
				"10.0.0.1",
				"10.255.255.255",
			},
			rejectIPs: []string{
				"11.0.0.1",
			},
		},
		{
			desc: "non-overlapping ranges stay separate",
			trustedIPs: []string{
				"10.0.0.0/24",
				"192.168.0.0/24",
			},
			expectedRanges: 2,
			passIPs: []string{
				"10.0.0.1",
				"192.168.0.1",
			},
			rejectIPs: []string{
				"172.16.0.1",
			},
		},
		{
			desc: "IPv6 overlapping ranges",
			trustedIPs: []string{
				"2001:db8::/48",
				"2001:db8::/64",
			},
			expectedRanges: 1,
			passIPs: []string{
				"2001:db8::1",
				"2001:db8:0:ffff:ffff:ffff:ffff:ffff",
			},
			rejectIPs: []string{
				"2001:db9::1",
			},
		},
		{
			desc: "mixed IPv4 and IPv6 ranges are not merged across families",
			trustedIPs: []string{
				"0.0.0.0/0",
				"2001:db8::/64",
			},
			expectedRanges: 2,
			passIPs: []string{
				"10.0.0.1",
				"255.255.255.255",
				"2001:db8::1",
			},
			rejectIPs: []string{
				"::1",
				"fe80::1",
				"2001:db9::1",
			},
		},
		{
			desc: "IPv6 /0 covers all IPv6 addresses",
			trustedIPs: []string{
				"::/0",
			},
			expectedRanges: 1,
			passIPs: []string{
				"::1",
				"2001:db8::1",
				"fe80::1",
				"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			},
			rejectIPs: []string{
				"10.0.0.1",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			checker, err := NewChecker(test.trustedIPs)
			require.NoError(t, err)
			require.Len(t, checker.authorizedRanges, test.expectedRanges)

			for _, ip := range test.passIPs {
				ok, err := checker.Contains(ip)
				require.NoError(t, err)
				assert.Truef(t, ok, "%s should have passed", ip)
			}

			for _, ip := range test.rejectIPs {
				ok, err := checker.Contains(ip)
				require.NoError(t, err)
				assert.Falsef(t, ok, "%s should have been rejected", ip)
			}
		})
	}
}

func TestContainsRangeBoundaries(t *testing.T) {
	trustedIPs := []string{
		"10.0.1.0/24",
		"10.0.3.0/24",
		"10.0.5.0/24",
	}

	checker, err := NewChecker(trustedIPs)
	require.NoError(t, err)

	testCases := []struct {
		desc       string
		addr       string
		authorized bool
	}{
		{
			desc:       "before first range",
			addr:       "10.0.0.255",
			authorized: false,
		},
		{
			desc:       "exact start of first range",
			addr:       "10.0.1.0",
			authorized: true,
		},
		{
			desc:       "middle of first range",
			addr:       "10.0.1.128",
			authorized: true,
		},
		{
			desc:       "exact end of first range",
			addr:       "10.0.1.255",
			authorized: true,
		},
		{
			desc:       "between first and second range",
			addr:       "10.0.2.0",
			authorized: false,
		},
		{
			desc:       "exact start of second range",
			addr:       "10.0.3.0",
			authorized: true,
		},
		{
			desc:       "exact end of second range",
			addr:       "10.0.3.255",
			authorized: true,
		},
		{
			desc:       "between second and third range",
			addr:       "10.0.4.128",
			authorized: false,
		},
		{
			desc:       "exact start of third range",
			addr:       "10.0.5.0",
			authorized: true,
		},
		{
			desc:       "exact end of third range",
			addr:       "10.0.5.255",
			authorized: true,
		},
		{
			desc:       "after last range",
			addr:       "10.0.6.0",
			authorized: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ok, err := checker.Contains(test.addr)
			require.NoError(t, err)
			assert.Equal(t, test.authorized, ok)
		})
	}
}

func TestContainsIP(t *testing.T) {
	trustedIPs := []string{"10.0.0.0/24", "2001:db8::/64"}
	checker, err := NewChecker(trustedIPs)
	require.NoError(t, err)

	testCases := []struct {
		desc       string
		ip         net.IP
		authorized bool
	}{
		{
			desc:       "IPv4 in range",
			ip:         net.ParseIP("10.0.0.50"),
			authorized: true,
		},
		{
			desc:       "IPv4 out of range",
			ip:         net.ParseIP("10.0.1.50"),
			authorized: false,
		},
		{
			desc:       "IPv6 in range",
			ip:         net.ParseIP("2001:db8::1"),
			authorized: true,
		},
		{
			desc:       "IPv6 out of range",
			ip:         net.ParseIP("2001:db9::1"),
			authorized: false,
		},
		{
			desc:       "IPv4-mapped IPv6 in range",
			ip:         net.ParseIP("::ffff:10.0.0.50"),
			authorized: true,
		},
		{
			desc:       "nil IP",
			ip:         nil,
			authorized: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := checker.ContainsIP(test.ip)
			assert.Equal(t, test.authorized, result)
		})
	}
}

func TestIPv4MappedIPv6(t *testing.T) {
	trustedIPs := []string{"192.168.1.0/24"}
	checker, err := NewChecker(trustedIPs)
	require.NoError(t, err)

	testCases := []struct {
		desc       string
		addr       string
		authorized bool
	}{
		{
			desc:       "IPv4-mapped IPv6 in range",
			addr:       "::ffff:192.168.1.50",
			authorized: true,
		},
		{
			desc:       "IPv4-mapped IPv6 out of range",
			addr:       "::ffff:192.168.2.50",
			authorized: false,
		},
		{
			desc:       "plain IPv4 in range",
			addr:       "192.168.1.50",
			authorized: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ok, err := checker.Contains(test.addr)
			require.NoError(t, err)
			assert.Equal(t, test.authorized, ok)
		})
	}
}

func TestContainsEmptyAddress(t *testing.T) {
	checker, err := NewChecker([]string{"10.0.0.0/24"})
	require.NoError(t, err)

	ok, err := checker.Contains("")
	require.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "empty IP address")
}

func TestIsAuthorizedVariations(t *testing.T) {
	trustedIPs := []string{
		"10.0.0.0/24",
		"2001:db8::/64",
	}
	checker, err := NewChecker(trustedIPs)
	require.NoError(t, err)

	testCases := []struct {
		desc       string
		addr       string
		authorized bool
	}{
		{
			desc:       "IP with port",
			addr:       "10.0.0.50:12345",
			authorized: true,
		},
		{
			desc:       "IP without port",
			addr:       "10.0.0.50",
			authorized: true,
		},
		{
			desc:       "IPv6 with brackets and port",
			addr:       "[2001:db8::1]:8080",
			authorized: true,
		},
		{
			desc:       "IPv6 without port",
			addr:       "2001:db8::1",
			authorized: true,
		},
		{
			desc:       "IPv4-mapped IPv6 with port",
			addr:       "[::ffff:10.0.0.50]:443",
			authorized: true,
		},
		{
			desc:       "IPv6 with zone ID",
			addr:       "fe80::1%eth0",
			authorized: false,
		},
		{
			desc:       "unauthorized IP with port",
			addr:       "192.168.1.1:80",
			authorized: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := checker.IsAuthorized(test.addr)
			if test.authorized {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestZoneIDHandling(t *testing.T) {
	t.Run("zone ID on single IP", func(t *testing.T) {
		checker, err := NewChecker([]string{"fe80::1"})
		require.NoError(t, err)

		ok, err := checker.Contains("fe80::1")
		require.NoError(t, err)
		assert.True(t, ok, "fe80::1 should match authorized fe80::1")

		ok, err = checker.Contains("fe80::1%eth0")
		require.NoError(t, err)
		assert.Truef(t, ok, "%s should match authorized fe80::1", "fe80::1%eth0")
	})

	t.Run("zone ID in range", func(t *testing.T) {
		checker, err := NewChecker([]string{"fe80::/16"})
		require.NoError(t, err)

		ok, err := checker.Contains("fe80::%eth0")
		require.NoError(t, err)
		assert.Truef(t, ok, "%s should match fe80::/16", "fe80::%eth0")

		ok, err = checker.Contains("fe80::1%eth0")
		require.NoError(t, err)
		assert.Truef(t, ok, "%s should match fe80::/16", "fe80::1%eth0")

		ok, err = checker.Contains("fe80:ffff:ffff:ffff:ffff:ffff:ffff:ffff%eth0")
		require.NoError(t, err)
		assert.Truef(t, ok, "%s should match fe80::/16 range end", "fe80:ffff:ffff:ffff:ffff:ffff:ffff:ffff%eth0")
	})

	t.Run("zone ID in config", func(t *testing.T) {
		checker, err := NewChecker([]string{"fe80::1%eth0"})
		require.NoError(t, err)

		ok, err := checker.Contains("fe80::1")
		require.NoError(t, err)
		assert.Truef(t, ok, "%s should match authorized %s", "fe80::1", "fe80::1%eth0")

		ok, err = checker.Contains("fe80::1%eth0")
		require.NoError(t, err)
		assert.Truef(t, ok, "%s should match authorized %s", "fe80::1%eth0", "fe80::1%eth0")
	})
}
