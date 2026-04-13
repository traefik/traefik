package xheaders

const (
	ForwardedProto             = "X-Forwarded-Proto"
	ForwardedFor               = "X-Forwarded-For"
	ForwardedHost              = "X-Forwarded-Host"
	ForwardedPort              = "X-Forwarded-Port"
	ForwardedServer            = "X-Forwarded-Server"
	ForwardedURI               = "X-Forwarded-Uri"
	ForwardedMethod            = "X-Forwarded-Method"
	ForwardedPrefix            = "X-Forwarded-Prefix"
	ForwardedTLSClientCert     = "X-Forwarded-Tls-Client-Cert"
	ForwardedTLSClientCertInfo = "X-Forwarded-Tls-Client-Cert-Info"
	RealIP                     = "X-Real-Ip"
)

// Headers contains all X-Forwarded header keys managed by Traefik.
var Headers = []string{
	ForwardedProto,
	ForwardedFor,
	ForwardedHost,
	ForwardedPort,
	ForwardedServer,
	ForwardedURI,
	ForwardedMethod,
	ForwardedPrefix,
	ForwardedTLSClientCert,
	ForwardedTLSClientCertInfo,
	RealIP,
}
