package loadbalancer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldSendSameSiteNone(t *testing.T) {
	testCases := []struct {
		desc      string
		userAgent string
		want      bool
	}{
		{
			desc:      "Compatible Chrome version",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36",
			want:      true,
		},
		{
			desc:      "Chrome 50",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36",
			want:      true,
		},
		{
			desc:      "Chrome 51",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36",
		},
		{
			desc:      "Chrome 66",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36",
		},
		{
			desc:      "Chrome 67",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Safari/537.36",
			want:      true,
		},
		{
			desc:      "Chromium 66",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chromium/66.0.3359.181 Safari/537.36",
		},
		{
			desc:      "Safari on iOS 12",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Mobile/15E148 Safari/604.1",
		},
		{
			desc:      "Safari on iOS 13",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 13_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0 Mobile/15E148 Safari/604.1",
			want:      true,
		},
		{
			desc:      "Safari on macOS 10.14",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Safari/605.1.15",
		},
		{
			desc:      "Safari on macOS 10.15",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Safari/605.1.15",
			want:      true,
		},
		{
			desc:      "Chrome on macOS 10.14",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36",
			want:      true,
		},
		{
			desc:      "Old UC Browser",
			userAgent: "Mozilla/5.0 (Linux; U; Android 9; en-US; Pixel Build/PQ3A.190801.002) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 UCBrowser/12.13.1.1200 Mobile Safari/537.36",
		},
		{
			desc:      "Supported UC Browser",
			userAgent: "Mozilla/5.0 (Linux; U; Android 9; en-US; Pixel Build/PQ3A.190801.002) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 UCBrowser/12.13.2.1200 Mobile Safari/537.36",
			want:      true,
		},
		{
			desc:      "Unknown client",
			userAgent: "unknown",
			want:      true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, shouldSendSameSiteNone(test.userAgent))
		})
	}
}
