package dynamic

// +k8s:deepcopy-gen=true

// Headers holds the headers middleware configuration.
// This middleware manages the requests and responses headers.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/headers/#customrequestheaders
type Headers struct {
	// RequestHeader defines header rules to modify the request
	RequestHeaders ModifyHeader `json:"requestHeaders,omitempty" toml:"requestHeaders,omitempty" yaml:"requestHeaders,omitempty" export:"true"`
	// ResponseHeader defines header rules to modify the request
	ResponseHeaders ModifyHeader `json:"responseHeaders,omitempty" toml:"responseHeaders,omitempty" yaml:"responseHeaders,omitempty" export:"true"`
	// SecurityHeader defines security rules to modify the request
	SecurityHeaders SecurityHeader `json:"securityHeader,omitempty" toml:"securityHeader,omitempty" yaml:"securityHeader,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

type ModifyHeader struct {
	Set    map[string]string `json:"set,omitempty" toml:"set,omitempty" yaml:"set,omitempty"  export:"true"`
	Append map[string]string `json:"append,omitempty" toml:"set,omitempty" yaml:"append,omitempty" export:"true"`
	Delete []string          `json:"delete,omitempty" toml:"delete,omitempty" yaml:"delete,omitempty" export:"true"`
}

func (mh *ModifyHeader) IsDefined() bool {
	return mh != nil && (len(mh.Set) != 0 || len(mh.Append) != 0 || len(mh.Delete) != 0)
}

// +k8s:deepcopy-gen=true

type SecurityHeader struct {
	// AccessControlAllowCredentials defines whether the request can include user credentials.
	AccessControlAllowCredentials bool `json:"accessControlAllowCredentials,omitempty" toml:"accessControlAllowCredentials,omitempty" yaml:"accessControlAllowCredentials,omitempty" export:"true"`
	// AccessControlAllowHeaders defines the Access-Control-Request-Headers values sent in preflight response.
	AccessControlAllowHeaders []string `json:"accessControlAllowHeaders,omitempty" toml:"accessControlAllowHeaders,omitempty" yaml:"accessControlAllowHeaders,omitempty" export:"true"`
	// AccessControlAllowMethods defines the Access-Control-Request-Method values sent in preflight response.
	AccessControlAllowMethods []string `json:"accessControlAllowMethods,omitempty" toml:"accessControlAllowMethods,omitempty" yaml:"accessControlAllowMethods,omitempty" export:"true"`
	// AccessControlAllowOriginList is a list of allowable origins. Can also be a wildcard origin "*".
	AccessControlAllowOriginList []string `json:"accessControlAllowOriginList,omitempty" toml:"accessControlAllowOriginList,omitempty" yaml:"accessControlAllowOriginList,omitempty"`
	// AccessControlAllowOriginListRegex is a list of allowable origins written following the Regular Expression syntax (https://golang.org/pkg/regexp/).
	AccessControlAllowOriginListRegex []string `json:"accessControlAllowOriginListRegex,omitempty" toml:"accessControlAllowOriginListRegex,omitempty" yaml:"accessControlAllowOriginListRegex,omitempty"`
	// AccessControlExposeHeaders defines the Access-Control-Expose-Headers values sent in preflight response.
	AccessControlExposeHeaders []string `json:"accessControlExposeHeaders,omitempty" toml:"accessControlExposeHeaders,omitempty" yaml:"accessControlExposeHeaders,omitempty" export:"true"`
	// AccessControlMaxAge defines the time that a preflight request may be cached.
	AccessControlMaxAge int64 `json:"accessControlMaxAge,omitempty" toml:"accessControlMaxAge,omitempty" yaml:"accessControlMaxAge,omitempty" export:"true"`
	// AddVaryHeader defines whether the Vary header is automatically added/updated when the AccessControlAllowOriginList is set.
	AddVaryHeader bool `json:"addVaryHeader,omitempty" toml:"addVaryHeader,omitempty" yaml:"addVaryHeader,omitempty" export:"true"`
	// AllowedHosts defines the fully qualified list of allowed domain names.
	AllowedHosts []string `json:"allowedHosts,omitempty" toml:"allowedHosts,omitempty" yaml:"allowedHosts,omitempty"`
	// HostsProxyHeaders defines the header keys that may hold a proxied hostname value for the request.
	HostsProxyHeaders []string `json:"hostsProxyHeaders,omitempty" toml:"hostsProxyHeaders,omitempty" yaml:"hostsProxyHeaders,omitempty" export:"true"`
	// SSLProxyHeaders defines the header keys with associated values that would indicate a valid HTTPS request.
	// It can be useful when using other proxies (example: "X-Forwarded-Proto": "https").
	SSLProxyHeaders map[string]string `json:"sslProxyHeaders,omitempty" toml:"sslProxyHeaders,omitempty" yaml:"sslProxyHeaders,omitempty"`
	// STSSeconds defines the max-age of the Strict-Transport-Security header.
	// If set to 0, the header is not set.
	STSSeconds int64 `json:"stsSeconds,omitempty" toml:"stsSeconds,omitempty" yaml:"stsSeconds,omitempty" export:"true"`
	// STSIncludeSubdomains defines whether the includeSubDomains directive is appended to the Strict-Transport-Security header.
	STSIncludeSubdomains bool `json:"stsIncludeSubdomains,omitempty" toml:"stsIncludeSubdomains,omitempty" yaml:"stsIncludeSubdomains,omitempty" export:"true"`
	// STSPreload defines whether the preload flag is appended to the Strict-Transport-Security header.
	STSPreload bool `json:"stsPreload,omitempty" toml:"stsPreload,omitempty" yaml:"stsPreload,omitempty" export:"true"`
	// ForceSTSHeader defines whether to add the STS header even when the connection is HTTP.
	ForceSTSHeader bool `json:"forceSTSHeader,omitempty" toml:"forceSTSHeader,omitempty" yaml:"forceSTSHeader,omitempty" export:"true"`
	// FrameDeny defines whether to add the X-Frame-Options header with the DENY value.
	FrameDeny bool `json:"frameDeny,omitempty" toml:"frameDeny,omitempty" yaml:"frameDeny,omitempty" export:"true"`
	// CustomFrameOptionsValue defines the X-Frame-Options header value.
	// This overrides the FrameDeny option.
	CustomFrameOptionsValue string `json:"customFrameOptionsValue,omitempty" toml:"customFrameOptionsValue,omitempty" yaml:"customFrameOptionsValue,omitempty"`
	// ContentTypeNosniff defines whether to add the X-Content-Type-Options header with the nosniff value.
	ContentTypeNosniff bool `json:"contentTypeNosniff,omitempty" toml:"contentTypeNosniff,omitempty" yaml:"contentTypeNosniff,omitempty" export:"true"`
	// BrowserXSSFilter defines whether to add the X-XSS-Protection header with the value 1; mode=block.
	BrowserXSSFilter bool `json:"browserXssFilter,omitempty" toml:"browserXssFilter,omitempty" yaml:"browserXssFilter,omitempty" export:"true"`
	// CustomBrowserXSSValue defines the X-XSS-Protection header value.
	// This overrides the BrowserXssFilter option.
	CustomBrowserXSSValue string `json:"customBrowserXSSValue,omitempty" toml:"customBrowserXSSValue,omitempty" yaml:"customBrowserXSSValue,omitempty"`
	// ContentSecurityPolicy defines the Content-Security-Policy header value.
	ContentSecurityPolicy string `json:"contentSecurityPolicy,omitempty" toml:"contentSecurityPolicy,omitempty" yaml:"contentSecurityPolicy,omitempty"`
	// PublicKey is the public key that implements HPKP to prevent MITM attacks with forged certificates.
	PublicKey string `json:"publicKey,omitempty" toml:"publicKey,omitempty" yaml:"publicKey,omitempty"`
	// ReferrerPolicy defines the Referrer-Policy header value.
	// This allows sites to control whether browsers forward the Referer header to other sites.
	ReferrerPolicy string `json:"referrerPolicy,omitempty" toml:"referrerPolicy,omitempty" yaml:"referrerPolicy,omitempty" export:"true"`
	// PermissionsPolicy defines the Permissions-Policy header value.
	// This allows sites to control browser features.
	PermissionsPolicy string `json:"permissionsPolicy,omitempty" toml:"permissionsPolicy,omitempty" yaml:"permissionsPolicy,omitempty" export:"true"`
	// IsDevelopment defines whether to mitigate the unwanted effects of the AllowedHosts, SSL, and STS options when developing.
	// Usually testing takes place using HTTP, not HTTPS, and on localhost, not your production domain.
	// If you would like your development environment to mimic production with complete Host blocking, SSL redirects,
	// and STS headers, leave this as false.
	IsDevelopment bool `json:"isDevelopment,omitempty" toml:"isDevelopment,omitempty" yaml:"isDevelopment,omitempty" export:"true"`
}

func (h *Headers) HasModifyHeadersDefined() bool {
	return h != nil && (h.RequestHeaders.IsDefined() || h.ResponseHeaders.IsDefined())
}

// HasCorsHeadersDefined checks to see if any of the cors header elements have been set.
func (sh *SecurityHeader) HasCorsHeadersDefined() bool {
	return sh != nil && (sh.AccessControlAllowCredentials ||
		len(sh.AccessControlAllowHeaders) != 0 ||
		len(sh.AccessControlAllowMethods) != 0 ||
		len(sh.AccessControlAllowOriginList) != 0 ||
		len(sh.AccessControlAllowOriginListRegex) != 0 ||
		len(sh.AccessControlExposeHeaders) != 0 ||
		sh.AccessControlMaxAge != 0 ||
		sh.AddVaryHeader)
}

// HasSecureHeadersDefined checks to see if any of the secure header elements have been set.
func (sh *SecurityHeader) HasSecureHeadersDefined() bool {
	return sh != nil && (len(sh.AllowedHosts) != 0 ||
		len(sh.HostsProxyHeaders) != 0 ||
		len(sh.SSLProxyHeaders) != 0 ||
		sh.STSSeconds != 0 ||
		sh.STSIncludeSubdomains ||
		sh.STSPreload ||
		sh.ForceSTSHeader ||
		sh.FrameDeny ||
		sh.CustomFrameOptionsValue != "" ||
		sh.ContentTypeNosniff ||
		sh.BrowserXSSFilter ||
		sh.CustomBrowserXSSValue != "" ||
		sh.ContentSecurityPolicy != "" ||
		sh.PublicKey != "" ||
		sh.ReferrerPolicy != "" ||
		sh.PermissionsPolicy != "" ||
		sh.IsDevelopment)
}
