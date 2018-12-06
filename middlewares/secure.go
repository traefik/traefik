package middlewares

import (
	"net/http"
	"strings"

	"github.com/containous/traefik/types"
	"github.com/unrolled/secure"
)

type SecureConfig struct {
	Secure *secure.Secure
}

// NewSecure constructs a new Secure instance with supplied options.
func NewSecure(headers *types.Headers) *SecureConfig {
	if headers == nil || !headers.HasSecureHeadersDefined() {
		return nil
	}

	return &SecureConfig{
		Secure: secure.New(secure.Options{
			AllowedHosts:            headers.AllowedHosts,
			HostsProxyHeaders:       headers.HostsProxyHeaders,
			SSLRedirect:             headers.SSLRedirect,
			SSLTemporaryRedirect:    headers.SSLTemporaryRedirect,
			SSLHost:                 headers.SSLHost,
			SSLForceHost:            headers.SSLForceHost,
			SSLProxyHeaders:         headers.SSLProxyHeaders,
			STSSeconds:              headers.STSSeconds,
			STSIncludeSubdomains:    headers.STSIncludeSubdomains,
			STSPreload:              headers.STSPreload,
			ForceSTSHeader:          headers.ForceSTSHeader,
			FrameDeny:               headers.FrameDeny,
			CustomFrameOptionsValue: headers.CustomFrameOptionsValue,
			ContentTypeNosniff:      headers.ContentTypeNosniff,
			BrowserXssFilter:        headers.BrowserXSSFilter,
			CustomBrowserXssValue:   headers.CustomBrowserXSSValue,
			ContentSecurityPolicy:   headers.ContentSecurityPolicy,
			PublicKey:               headers.PublicKey,
			ReferrerPolicy:          headers.ReferrerPolicy,
			IsDevelopment:           headers.IsDevelopment,
		}),
	}
}

// HandlerFuncWithNextForRequestOnlyWithContextCheck is a special wrapper for Traefik that checks the context for stripped/modified paths.
// Note that this is for requests only and will not write any headers.
func (s *SecureConfig) HandlerFuncWithNextForRequestOnlyWithContextCheck(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	requestURL := r.URL

	if stripPrefix, stripPrefixOk := r.Context().Value(StripPrefixKey).(string); stripPrefixOk {
		if len(stripPrefix) > 0 {
			requestURL.Path = stripPrefix
		}
	}

	if addPrefix, addPrefixOk := r.Context().Value(AddPrefixKey).(string); addPrefixOk {
		if len(addPrefix) > 0 {
			requestURL.Path = strings.Replace(requestURL.Path, addPrefix, "", 1)
		}
	}

	if replacePath, replacePathOk := r.Context().Value(ReplacePathKey).(string); replacePathOk {
		if len(replacePath) > 0 {
			requestURL.Path = replacePath
		}
	}

	r.URL = requestURL
	// make sure the request URI corresponds the rewritten URL
	r.RequestURI = r.URL.RequestURI()

	s.Secure.HandlerFuncWithNextForRequestOnly(w, r, next)
}
