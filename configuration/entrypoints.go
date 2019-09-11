package configuration

import (
	"fmt"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
)

// EntryPoint holds an entry point configuration of the reverse proxy (ip, port, TLS...)
type EntryPoint struct {
	Address              string
	TLS                  *tls.TLS          `export:"true"`
	Redirect             *types.Redirect   `export:"true"`
	Auth                 *types.Auth       `export:"true"`
	WhitelistSourceRange []string          // Deprecated
	WhiteList            *types.WhiteList  `export:"true"`
	Compress             bool              `export:"true"`
	ProxyProtocol        *ProxyProtocol    `export:"true"`
	ForwardedHeaders     *ForwardedHeaders `export:"true"`
}

// ProxyProtocol contains Proxy-Protocol configuration
type ProxyProtocol struct {
	Insecure   bool `export:"true"`
	TrustedIPs []string
}

// ForwardedHeaders Trust client forwarding headers
type ForwardedHeaders struct {
	Insecure   bool `export:"true"`
	TrustedIPs []string
}

// EntryPoints holds entry points configuration of the reverse proxy (ip, port, TLS...)
type EntryPoints map[string]*EntryPoint

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (ep EntryPoints) String() string {
	return fmt.Sprintf("%+v", map[string]*EntryPoint(ep))
}

// Get return the EntryPoints map
func (ep *EntryPoints) Get() interface{} {
	return *ep
}

// SetValue sets the EntryPoints map with val
func (ep *EntryPoints) SetValue(val interface{}) {
	*ep = val.(EntryPoints)
}

// Type is type of the struct
func (ep *EntryPoints) Type() string {
	return "entrypoints"
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (ep *EntryPoints) Set(value string) error {
	result := parseEntryPointsConfiguration(value)

	var whiteListSourceRange []string
	if len(result["whitelistsourcerange"]) > 0 {
		whiteListSourceRange = strings.Split(result["whitelistsourcerange"], ",")
	}

	compress := toBool(result, "compress")

	configTLS, err := makeEntryPointTLS(result)
	if err != nil {
		return err
	}

	(*ep)[result["name"]] = &EntryPoint{
		Address:              result["address"],
		TLS:                  configTLS,
		Auth:                 makeEntryPointAuth(result),
		Redirect:             makeEntryPointRedirect(result),
		Compress:             compress,
		WhitelistSourceRange: whiteListSourceRange,
		WhiteList:            makeWhiteList(result),
		ProxyProtocol:        makeEntryPointProxyProtocol(result),
		ForwardedHeaders:     makeEntryPointForwardedHeaders(result),
	}

	return nil
}

func makeWhiteList(result map[string]string) *types.WhiteList {
	var wl *types.WhiteList
	if rawRange, ok := result["whitelist_sourcerange"]; ok {
		wl = &types.WhiteList{
			SourceRange:      strings.Split(rawRange, ","),
			UseXForwardedFor: toBool(result, "whitelist_usexforwardedfor"),
		}
	}
	return wl
}

func makeEntryPointAuth(result map[string]string) *types.Auth {
	var basic *types.Basic
	if v, ok := result["auth_basic_users"]; ok {
		basic = &types.Basic{
			Users:        strings.Split(v, ","),
			RemoveHeader: toBool(result, "auth_basic_removeheader"),
		}
	}

	var digest *types.Digest
	if v, ok := result["auth_digest_users"]; ok {
		digest = &types.Digest{
			Users:        strings.Split(v, ","),
			RemoveHeader: toBool(result, "auth_digest_removeheader"),
		}
	}

	var forward *types.Forward
	if address, ok := result["auth_forward_address"]; ok {
		var clientTLS *types.ClientTLS

		cert := result["auth_forward_tls_cert"]
		key := result["auth_forward_tls_key"]
		insecureSkipVerify := toBool(result, "auth_forward_tls_insecureskipverify")

		if len(cert) > 0 && len(key) > 0 || insecureSkipVerify {
			clientTLS = &types.ClientTLS{
				CA:                 result["auth_forward_tls_ca"],
				CAOptional:         toBool(result, "auth_forward_tls_caoptional"),
				Cert:               cert,
				Key:                key,
				InsecureSkipVerify: insecureSkipVerify,
			}
		}

		var authResponseHeaders []string
		if v, ok := result["auth_forward_authresponseheaders"]; ok {
			authResponseHeaders = strings.Split(v, ",")
		}

		forward = &types.Forward{
			Address:             address,
			TLS:                 clientTLS,
			TrustForwardHeader:  toBool(result, "auth_forward_trustforwardheader"),
			AuthResponseHeaders: authResponseHeaders,
		}
	}

	var auth *types.Auth
	if basic != nil || digest != nil || forward != nil {
		auth = &types.Auth{
			Basic:       basic,
			Digest:      digest,
			Forward:     forward,
			HeaderField: result["auth_headerfield"],
		}
	}

	return auth
}

func makeEntryPointProxyProtocol(result map[string]string) *ProxyProtocol {
	var proxyProtocol *ProxyProtocol

	ppTrustedIPs := result["proxyprotocol_trustedips"]
	if len(result["proxyprotocol_insecure"]) > 0 || len(ppTrustedIPs) > 0 {
		proxyProtocol = &ProxyProtocol{
			Insecure: toBool(result, "proxyprotocol_insecure"),
		}
		if len(ppTrustedIPs) > 0 {
			proxyProtocol.TrustedIPs = strings.Split(ppTrustedIPs, ",")
		}
	}

	if proxyProtocol != nil && proxyProtocol.Insecure {
		log.Warn("ProxyProtocol.Insecure:true is dangerous. Please use 'ProxyProtocol.TrustedIPs:IPs' and remove 'ProxyProtocol.Insecure:true'")
	}

	return proxyProtocol
}

func makeEntryPointForwardedHeaders(result map[string]string) *ForwardedHeaders {
	// TODO must be changed to false by default in the next breaking version.
	forwardedHeaders := &ForwardedHeaders{Insecure: true}
	if _, ok := result["forwardedheaders_insecure"]; ok {
		forwardedHeaders.Insecure = toBool(result, "forwardedheaders_insecure")
	}

	fhTrustedIPs := result["forwardedheaders_trustedips"]
	if len(fhTrustedIPs) > 0 {
		// TODO must be removed in the next breaking version.
		forwardedHeaders.Insecure = toBool(result, "forwardedheaders_insecure")
		forwardedHeaders.TrustedIPs = strings.Split(fhTrustedIPs, ",")
	}

	return forwardedHeaders
}

func makeEntryPointRedirect(result map[string]string) *types.Redirect {
	var redirect *types.Redirect

	if len(result["redirect_entrypoint"]) > 0 || len(result["redirect_regex"]) > 0 || len(result["redirect_replacement"]) > 0 {
		redirect = &types.Redirect{
			EntryPoint:  result["redirect_entrypoint"],
			Regex:       result["redirect_regex"],
			Replacement: result["redirect_replacement"],
			Permanent:   toBool(result, "redirect_permanent"),
		}
	}

	return redirect
}

func makeEntryPointTLS(result map[string]string) (*tls.TLS, error) {
	var configTLS *tls.TLS

	if len(result["tls"]) > 0 {
		certs := tls.Certificates{}
		if err := certs.Set(result["tls"]); err != nil {
			return nil, err
		}
		configTLS = &tls.TLS{
			Certificates: certs,
		}
	} else if len(result["tls_acme"]) > 0 {
		configTLS = &tls.TLS{
			Certificates: tls.Certificates{},
		}
	}

	if configTLS != nil {
		if len(result["ca"]) > 0 {
			files := tls.FilesOrContents{}
			files.Set(result["ca"])
			optional := toBool(result, "ca_optional")
			configTLS.ClientCA = tls.ClientCA{
				Files:    files,
				Optional: optional,
			}
		}

		if len(result["tls_minversion"]) > 0 {
			configTLS.MinVersion = result["tls_minversion"]
		}

		if len(result["tls_ciphersuites"]) > 0 {
			configTLS.CipherSuites = strings.Split(result["tls_ciphersuites"], ",")
		}

		if len(result["tls_snistrict"]) > 0 {
			configTLS.SniStrict = toBool(result, "tls_snistrict")
		}

		if len(result["tls_defaultcertificate_cert"]) > 0 && len(result["tls_defaultcertificate_key"]) > 0 {
			configTLS.DefaultCertificate = &tls.Certificate{
				CertFile: tls.FileOrContent(result["tls_defaultcertificate_cert"]),
				KeyFile:  tls.FileOrContent(result["tls_defaultcertificate_key"]),
			}
		}
	}

	return configTLS, nil
}

func parseEntryPointsConfiguration(raw string) map[string]string {
	sections := strings.Fields(raw)

	config := make(map[string]string)
	for _, part := range sections {
		field := strings.SplitN(part, ":", 2)
		name := strings.ToLower(strings.Replace(field[0], ".", "_", -1))
		if len(field) > 1 {
			config[name] = field[1]
		} else {
			if strings.EqualFold(name, "TLS") {
				config["tls_acme"] = "TLS"
			} else {
				config[name] = ""
			}
		}
	}
	return config
}

func toBool(conf map[string]string, key string) bool {
	if val, ok := conf[key]; ok {
		return strings.EqualFold(val, "true") ||
			strings.EqualFold(val, "enable") ||
			strings.EqualFold(val, "on")
	}
	return false
}
