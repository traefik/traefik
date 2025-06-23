package ingressnginx

import (
	"errors"
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
	AuthResponseHeaders *string `annotation:"nginx.ingress.kubernetes.io/auth-response-headers"`

	ForceSSLRedirect *bool `annotation:"nginx.ingress.kubernetes.io/force-ssl-redirect"`
	SSLRedirect      *bool `annotation:"nginx.ingress.kubernetes.io/ssl-redirect"`

	SSLPassthrough *bool `annotation:"nginx.ingress.kubernetes.io/ssl-passthrough"`

	UseRegex *bool `annotation:"nginx.ingress.kubernetes.io/use-regex"`

	Affinity              *string `annotation:"nginx.ingress.kubernetes.io/affinity"`
	SessionCookieName     *string `annotation:"nginx.ingress.kubernetes.io/session-cookie-name"`
	SessionCookieSecure   *bool   `annotation:"nginx.ingress.kubernetes.io/session-cookie-secure"`
	SessionCookiePath     *string `annotation:"nginx.ingress.kubernetes.io/session-cookie-path"`
	SessionCookieDomain   *string `annotation:"nginx.ingress.kubernetes.io/session-cookie-domain"`
	SessionCookieSameSite *string `annotation:"nginx.ingress.kubernetes.io/session-cookie-samesite"`
	SessionCookieMaxAge   *int    `annotation:"nginx.ingress.kubernetes.io/session-cookie-max-age"`

	ServiceUpstream *bool `annotation:"nginx.ingress.kubernetes.io/service-upstream"`

	BackendProtocol *string `annotation:"nginx.ingress.kubernetes.io/backend-protocol"`

	ProxySSLSecret     *string `annotation:"nginx.ingress.kubernetes.io/proxy-ssl-secret"`
	ProxySSLVerify     *string `annotation:"nginx.ingress.kubernetes.io/proxy-ssl-verify"`
	ProxySSLName       *string `annotation:"nginx.ingress.kubernetes.io/proxy-ssl-name"`
	ProxySSLServerName *string `annotation:"nginx.ingress.kubernetes.io/proxy-ssl-server-name"`

	EnableCORS                 *bool     `annotation:"nginx.ingress.kubernetes.io/enable-cors"`
	EnableCORSAllowCredentials *bool     `annotation:"nginx.ingress.kubernetes.io/cors-allow-credentials"`
	CORSExposeHeaders          *[]string `annotation:"nginx.ingress.kubernetes.io/cors-expose-headers"`
	CORSAllowHeaders           *[]string `annotation:"nginx.ingress.kubernetes.io/cors-allow-headers"`
	CORSAllowMethods           *[]string `annotation:"nginx.ingress.kubernetes.io/cors-allow-methods"`
	CORSAllowOrigin            *[]string `annotation:"nginx.ingress.kubernetes.io/cors-allow-origin"`
	CORSMaxAge                 *int      `annotation:"nginx.ingress.kubernetes.io/cors-max-age"`
}

// parseIngressConfig parses the annotations from an Ingress object into an ingressConfig struct.
func parseIngressConfig(ing *netv1.Ingress) (ingressConfig, error) {
	cfg := ingressConfig{}
	cfgType := reflect.TypeOf(cfg)
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
			parsed, err := strconv.ParseBool(val)
			if err == nil {
				cfgValue.Field(i).Set(reflect.ValueOf(&parsed))
			}
		case reflect.Int:
			parsed, err := strconv.Atoi(val)
			if err == nil {
				cfgValue.Field(i).Set(reflect.ValueOf(&parsed))
			}
		case reflect.Slice:
			if field.Type.Elem().Elem().Kind() == reflect.String {
				// Handle slice of strings
				var slice []string
				elements := strings.Split(val, ",")
				for _, elt := range elements {
					slice = append(slice, strings.TrimSpace(elt))
				}
				cfgValue.Field(i).Set(reflect.ValueOf(&slice))
			} else {
				return cfg, errors.New("unsupported slice type in annotations")
			}
		default:
			return cfg, errors.New("unsupported kind")
		}
	}

	return cfg, nil
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
