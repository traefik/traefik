package responsemodifiers

import (
	"net/http"

	"github.com/containous/traefik/config"
	"github.com/unrolled/secure"
)

func buildHeaders(headers *config.Headers) func(*http.Response) error {
	opt := secure.Options{
		BrowserXssFilter:        headers.BrowserXSSFilter,
		ContentTypeNosniff:      headers.ContentTypeNosniff,
		ForceSTSHeader:          headers.ForceSTSHeader,
		FrameDeny:               headers.FrameDeny,
		IsDevelopment:           headers.IsDevelopment,
		SSLRedirect:             headers.SSLRedirect,
		SSLForceHost:            headers.SSLForceHost,
		SSLTemporaryRedirect:    headers.SSLTemporaryRedirect,
		STSIncludeSubdomains:    headers.STSIncludeSubdomains,
		STSPreload:              headers.STSPreload,
		ContentSecurityPolicy:   headers.ContentSecurityPolicy,
		CustomBrowserXssValue:   headers.CustomBrowserXSSValue,
		CustomFrameOptionsValue: headers.CustomFrameOptionsValue,
		PublicKey:               headers.PublicKey,
		ReferrerPolicy:          headers.ReferrerPolicy,
		SSLHost:                 headers.SSLHost,
		AllowedHosts:            headers.AllowedHosts,
		HostsProxyHeaders:       headers.HostsProxyHeaders,
		SSLProxyHeaders:         headers.SSLProxyHeaders,
		STSSeconds:              headers.STSSeconds,
	}

	return func(resp *http.Response) error {
		if headers.HasCustomHeadersDefined() {
			// Loop through Custom response headers
			for header, value := range headers.CustomResponseHeaders {
				if value == "" {
					resp.Header.Del(header)
				} else {
					resp.Header.Set(header, value)
				}
			}
		}

		if headers.HasSecureHeadersDefined() {
			err := secure.New(opt).ModifyResponseHeaders(resp)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
