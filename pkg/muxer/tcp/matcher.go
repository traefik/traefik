package tcp

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/ip"
)

var tcpFuncs = map[string]func(*matchersTree, ...string) error{
	"ALPN":          expect1Parameter(alpn),
	"ClientIP":      expect1Parameter(clientIP),
	"HostSNI":       expect1Parameter(hostSNI),
	"HostSNIRegexp": expect1Parameter(hostSNIRegexp),
}

func expect1Parameter(fn func(*matchersTree, ...string) error) func(*matchersTree, ...string) error {
	return func(route *matchersTree, s ...string) error {
		if len(s) != 1 {
			return fmt.Errorf("unexpected number of parameters; got %d, expected 1", len(s))
		}

		return fn(route, s...)
	}
}

// alpn checks if any of the connection ALPN protocols matches one of the matcher protocols.
func alpn(tree *matchersTree, protos ...string) error {
	proto := protos[0]

	if proto == tlsalpn01.ACMETLS1Protocol {
		return fmt.Errorf("invalid protocol value for ALPN matcher, %q is not allowed", proto)
	}

	tree.matcher = func(meta ConnData) bool {
		for _, alpnProto := range meta.alpnProtos {
			if alpnProto == proto {
				return true
			}
		}

		return false
	}

	return nil
}

func clientIP(tree *matchersTree, clientIP ...string) error {
	checker, err := ip.NewChecker(clientIP)
	if err != nil {
		return fmt.Errorf("initializing IP checker for ClientIP matcher: %w", err)
	}

	tree.matcher = func(meta ConnData) bool {
		ok, err := checker.Contains(meta.remoteIP)
		if err != nil {
			log.Warn().Err(err).Msg("ClientIP matcher: could not match remote address")
			return false
		}
		return ok
	}

	return nil
}

var almostFQDN = regexp.MustCompile(`^[[:alnum:]\.-]+$`)

// hostSNI checks if the SNI Host of the connection match the matcher host.
func hostSNI(tree *matchersTree, hosts ...string) error {
	host := hosts[0]

	if host == "*" {
		// Since a HostSNI(`*`) rule has been provided as catchAll for non-TLS TCP,
		// it allows matching with an empty serverName.
		tree.matcher = func(meta ConnData) bool { return true }
		return nil
	}

	if !almostFQDN.MatchString(host) {
		return fmt.Errorf("invalid value for HostSNI matcher, %q is not a valid hostname", host)
	}

	tree.matcher = func(meta ConnData) bool {
		if meta.serverName == "" {
			return false
		}

		if host == meta.serverName {
			return true
		}

		// trim trailing period in case of FQDN
		host = strings.TrimSuffix(host, ".")

		return host == meta.serverName
	}

	return nil
}

// hostSNIRegexp checks if the SNI Host of the connection matches the matcher host regexp.
func hostSNIRegexp(tree *matchersTree, templates ...string) error {
	template := templates[0]

	if !isASCII(template) {
		return fmt.Errorf("invalid value for HostSNIRegexp matcher, %q is not a valid hostname", template)
	}

	re, err := regexp.Compile(template)
	if err != nil {
		return fmt.Errorf("compiling HostSNIRegexp matcher: %w", err)
	}

	tree.matcher = func(meta ConnData) bool {
		return re.MatchString(meta.serverName)
	}

	return nil
}

// isASCII checks if the given string contains only ASCII characters.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}

	return true
}
