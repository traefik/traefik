package ingressnginx

import (
	"reflect"
	"strconv"
	"strings"

	netv1 "k8s.io/api/networking/v1"
)

type ingressConfig struct {
	AuthType       *string `annotation:"nginx.ingress.kubernetes.io/auth-type"`
	AuthSecret     *string `annotation:"nginx.ingress.kubernetes.io/auth-secret"`
	AuthRealm      *string `annotation:"nginx.ingress.kubernetes.io/auth-realm"`
	AuthSecretType *string `annotation:"nginx.ingress.kubernetes.io/auth-secret-type"`

	AuthURL             *string `annotation:"nginx.ingress.kubernetes.io/auth-url"`
	AuthSignin          *string `annotation:"nginx.ingress.kubernetes.io/auth-signin"`
	AuthResponseHeaders *string `annotation:"nginx.ingress.kubernetes.io/auth-response-headers"`

	AuthTLSSecret                    *string `annotation:"nginx.ingress.kubernetes.io/auth-tls-secret"`
	AuthTLSVerifyClient              *string `annotation:"nginx.ingress.kubernetes.io/auth-tls-verify-client"`
	AuthTLSPassCertificateToUpstream *bool   `annotation:"nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream"`

	ForceSSLRedirect *bool `annotation:"nginx.ingress.kubernetes.io/force-ssl-redirect"`
	SSLRedirect      *bool `annotation:"nginx.ingress.kubernetes.io/ssl-redirect"`

	SSLPassthrough *bool `annotation:"nginx.ingress.kubernetes.io/ssl-passthrough"`

	UseRegex      *bool   `annotation:"nginx.ingress.kubernetes.io/use-regex"`
	RewriteTarget *string `annotation:"nginx.ingress.kubernetes.io/rewrite-target"`
	AppRoot       *string `annotation:"nginx.ingress.kubernetes.io/app-root"`

	PermanentRedirect     *string `annotation:"nginx.ingress.kubernetes.io/permanent-redirect"`
	PermanentRedirectCode *int    `annotation:"nginx.ingress.kubernetes.io/permanent-redirect-code"`
	TemporalRedirect      *string `annotation:"nginx.ingress.kubernetes.io/temporal-redirect"`
	TemporalRedirectCode  *int    `annotation:"nginx.ingress.kubernetes.io/temporal-redirect-code"`

	FromToWwwRedirect *bool `annotation:"nginx.ingress.kubernetes.io/from-to-www-redirect"`

	Affinity              *string `annotation:"nginx.ingress.kubernetes.io/affinity"`
	SessionCookieName     *string `annotation:"nginx.ingress.kubernetes.io/session-cookie-name"`
	SessionCookieSecure   *bool   `annotation:"nginx.ingress.kubernetes.io/session-cookie-secure"`
	SessionCookiePath     *string `annotation:"nginx.ingress.kubernetes.io/session-cookie-path"`
	SessionCookieDomain   *string `annotation:"nginx.ingress.kubernetes.io/session-cookie-domain"`
	SessionCookieSameSite *string `annotation:"nginx.ingress.kubernetes.io/session-cookie-samesite"`
	SessionCookieMaxAge   *int    `annotation:"nginx.ingress.kubernetes.io/session-cookie-max-age"`
	SessionCookieExpires  *int    `annotation:"nginx.ingress.kubernetes.io/session-cookie-expires"`

	ServiceUpstream *bool `annotation:"nginx.ingress.kubernetes.io/service-upstream"`

	BackendProtocol *string `annotation:"nginx.ingress.kubernetes.io/backend-protocol"`

	ProxySSLSecret      *string `annotation:"nginx.ingress.kubernetes.io/proxy-ssl-secret"`
	ProxySSLVerify      *string `annotation:"nginx.ingress.kubernetes.io/proxy-ssl-verify"`
	ProxySSLName        *string `annotation:"nginx.ingress.kubernetes.io/proxy-ssl-name"`
	ProxySSLServerName  *string `annotation:"nginx.ingress.kubernetes.io/proxy-ssl-server-name"`
	ProxyConnectTimeout *int    `annotation:"nginx.ingress.kubernetes.io/proxy-connect-timeout"`

	EnableCORS                 *bool     `annotation:"nginx.ingress.kubernetes.io/enable-cors"`
	EnableCORSAllowCredentials *bool     `annotation:"nginx.ingress.kubernetes.io/cors-allow-credentials"`
	CORSExposeHeaders          *[]string `annotation:"nginx.ingress.kubernetes.io/cors-expose-headers"`
	CORSAllowHeaders           *[]string `annotation:"nginx.ingress.kubernetes.io/cors-allow-headers"`
	CORSAllowMethods           *[]string `annotation:"nginx.ingress.kubernetes.io/cors-allow-methods"`
	CORSAllowOrigin            *[]string `annotation:"nginx.ingress.kubernetes.io/cors-allow-origin"`
	CORSMaxAge                 *int      `annotation:"nginx.ingress.kubernetes.io/cors-max-age"`

	WhitelistSourceRange *string `annotation:"nginx.ingress.kubernetes.io/whitelist-source-range"`
	AllowlistSourceRange *string `annotation:"nginx.ingress.kubernetes.io/allowlist-source-range"`

	CustomHeaders *string `annotation:"nginx.ingress.kubernetes.io/custom-headers"`
	UpstreamVhost *string `annotation:"nginx.ingress.kubernetes.io/upstream-vhost"`

	// ProxyRequestBuffering controls whether request buffering is enabled.
	ProxyRequestBuffering *string `annotation:"nginx.ingress.kubernetes.io/proxy-request-buffering"`
	// ClientBodyBufferSize sets the size of the buffer used for reading request body.
	ClientBodyBufferSize *string `annotation:"nginx.ingress.kubernetes.io/client-body-buffer-size"`
	// ProxyBodySize sets the maximum allowed size of the client request body.
	ProxyBodySize *string `annotation:"nginx.ingress.kubernetes.io/proxy-body-size"`

	// ProxyBuffering controls whether response buffering is enabled.
	ProxyBuffering *string `annotation:"nginx.ingress.kubernetes.io/proxy-buffering"`
	// ProxyBufferSize sets the size of the memory buffer used for reading the response.
	ProxyBufferSize *string `annotation:"nginx.ingress.kubernetes.io/proxy-buffer-size"`
	// ProxyBuffersNumber sets the number of memory buffers used for reading the response.
	ProxyBuffersNumber *int `annotation:"nginx.ingress.kubernetes.io/proxy-buffers-number"`
	// ProxyMaxTempFileSize sets the maximum size of a temporary file used to buffer responses.
	ProxyMaxTempFileSize *string `annotation:"nginx.ingress.kubernetes.io/proxy-max-temp-file-size"`
}

// parseIngressConfig parses the annotations from an Ingress object into an ingressConfig struct.
func parseIngressConfig(ing *netv1.Ingress) ingressConfig {
	cfg := ingressConfig{}
	cfgType := reflect.TypeFor[ingressConfig]()
	cfgValue := reflect.ValueOf(&cfg).Elem()

	for i := range cfgType.NumField() {
		field := cfgType.Field(i)
		annotation := field.Tag.Get("annotation")
		if annotation == "" {
			continue
		}

		val, ok := ing.GetAnnotations()[annotation]
		if !ok {
			continue
		}

		switch field.Type.Elem().Kind() {
		case reflect.String:
			cfgValue.Field(i).Set(reflect.ValueOf(&val))
		case reflect.Bool:
			b := strings.EqualFold(val, "true")
			cfgValue.Field(i).Set(reflect.ValueOf(&b))
		case reflect.Int:
			parsed, err := strconv.Atoi(val)
			if err == nil {
				cfgValue.Field(i).Set(reflect.ValueOf(&parsed))
			}
		case reflect.Slice:
			if field.Type.Elem().Elem().Kind() == reflect.String {
				// Handle slice of strings
				var slice []string
				for elt := range strings.SplitSeq(val, ",") {
					slice = append(slice, strings.TrimSpace(elt))
				}
				cfgValue.Field(i).Set(reflect.ValueOf(&slice))
			}
		default:
			continue
		}
	}

	return cfg
}

// parseBackendProtocol parses the backend protocol annotation and returns the corresponding protocol string.
func parseBackendProtocol(bp string) string {
	switch strings.ToUpper(bp) {
	case "HTTPS", "GRPCS":
		return "https"
	case "GRPC":
		return "h2c"
	default:
		return "http"
	}
}
