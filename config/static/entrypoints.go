package static

import (
	"fmt"
	"strings"

	"github.com/containous/traefik/log"
)

// EntryPoint holds the entry point configuration.
type EntryPoint struct {
	Address          string
	Transport        *EntryPointsTransport
	ProxyProtocol    *ProxyProtocol
	ForwardedHeaders *ForwardedHeaders
}

// ForwardedHeaders Trust client forwarding headers.
type ForwardedHeaders struct {
	Insecure   bool
	TrustedIPs []string
}

// ProxyProtocol contains Proxy-Protocol configuration.
type ProxyProtocol struct {
	Insecure   bool `export:"true"`
	TrustedIPs []string
}

// EntryPoints holds the HTTP entry point list.
type EntryPoints map[string]*EntryPoint

// EntryPointsTransport configures communication between clients and Traefik.
type EntryPointsTransport struct {
	LifeCycle          *LifeCycle          `description:"Timeouts influencing the server life cycle" export:"true"`
	RespondingTimeouts *RespondingTimeouts `description:"Timeouts for incoming requests to the Traefik instance" export:"true"`
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (ep EntryPoints) String() string {
	return fmt.Sprintf("%+v", map[string]*EntryPoint(ep))
}

// Get return the EntryPoints map.
func (ep *EntryPoints) Get() interface{} {
	return *ep
}

// SetValue sets the EntryPoints map with val.
func (ep *EntryPoints) SetValue(val interface{}) {
	*ep = val.(EntryPoints)
}

// Type is type of the struct.
func (ep *EntryPoints) Type() string {
	return "entrypoints"
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (ep *EntryPoints) Set(value string) error {
	result := parseEntryPointsConfiguration(value)

	(*ep)[result["name"]] = &EntryPoint{
		Address:          result["address"],
		ProxyProtocol:    makeEntryPointProxyProtocol(result),
		ForwardedHeaders: makeEntryPointForwardedHeaders(result),
	}

	return nil
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
		log.Warn("ProxyProtocol.insecure:true is dangerous. Please use 'ProxyProtocol.TrustedIPs:IPs' and remove 'ProxyProtocol.insecure:true'")
	}

	return proxyProtocol
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

func makeEntryPointForwardedHeaders(result map[string]string) *ForwardedHeaders {
	forwardedHeaders := &ForwardedHeaders{}
	forwardedHeaders.Insecure = toBool(result, "forwardedheaders_insecure")

	fhTrustedIPs := result["forwardedheaders_trustedips"]
	if len(fhTrustedIPs) > 0 {
		forwardedHeaders.TrustedIPs = strings.Split(fhTrustedIPs, ",")
	}

	return forwardedHeaders
}
