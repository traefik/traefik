package tcp

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/ip"
)

var tcpFuncsV2 = map[string]func(*matchersTree, ...string) error{
	"ALPN":          alpnV2,
	"ClientIP":      clientIPV2,
	"HostSNI":       hostSNIV2,
	"HostSNIRegexp": hostSNIRegexpV2,
}

func clientIPV2(tree *matchersTree, clientIPs ...string) error {
	checker, err := ip.NewChecker(clientIPs)
	if err != nil {
		return fmt.Errorf("could not initialize IP Checker for \"ClientIP\" matcher: %w", err)
	}

	tree.matcher = func(meta ConnData) bool {
		if meta.remoteIP == "" {
			return false
		}

		ok, err := checker.Contains(meta.remoteIP)
		if err != nil {
			log.Warn().Err(err).Msg("ClientIP matcher: could not match remote address")
			return false
		}
		return ok
	}

	return nil
}

// alpnV2 checks if any of the connection ALPN protocols matches one of the matcher protocols.
func alpnV2(tree *matchersTree, protos ...string) error {
	if len(protos) == 0 {
		return errors.New("empty value for \"ALPN\" matcher is not allowed")
	}

	for _, proto := range protos {
		if proto == tlsalpn01.ACMETLS1Protocol {
			return fmt.Errorf("invalid protocol value for \"ALPN\" matcher, %q is not allowed", proto)
		}
	}

	tree.matcher = func(meta ConnData) bool {
		for _, proto := range meta.alpnProtos {
			for _, filter := range protos {
				if proto == filter {
					return true
				}
			}
		}

		return false
	}

	return nil
}

// hostSNIV2 checks if the SNI Host of the connection match the matcher host.
func hostSNIV2(tree *matchersTree, hosts ...string) error {
	if len(hosts) == 0 {
		return errors.New("empty value for \"HostSNI\" matcher is not allowed")
	}

	for i, host := range hosts {
		// Special case to allow global wildcard
		if host == "*" {
			continue
		}

		if !hostOrIP.MatchString(host) {
			return fmt.Errorf("invalid value for \"HostSNI\" matcher, %q is not a valid hostname or IP", host)
		}

		hosts[i] = strings.ToLower(host)
	}

	tree.matcher = func(meta ConnData) bool {
		// Since a HostSNI(`*`) rule has been provided as catchAll for non-TLS TCP,
		// it allows matching with an empty serverName.
		// Which is why we make sure to take that case into account before
		// checking meta.serverName.
		if hosts[0] == "*" {
			return true
		}

		if meta.serverName == "" {
			return false
		}

		for _, host := range hosts {
			if host == "*" {
				return true
			}

			if host == meta.serverName {
				return true
			}

			// trim trailing period in case of FQDN
			host = strings.TrimSuffix(host, ".")
			if host == meta.serverName {
				return true
			}
		}

		return false
	}

	return nil
}

// hostSNIRegexpV2 checks if the SNI Host of the connection matches the matcher host regexp.
func hostSNIRegexpV2(tree *matchersTree, templates ...string) error {
	if len(templates) == 0 {
		return errors.New("empty value for \"HostSNIRegexp\" matcher is not allowed")
	}

	var regexps []*regexp.Regexp

	for _, template := range templates {
		preparedPattern, err := preparePattern(template)
		if err != nil {
			return fmt.Errorf("invalid pattern value for \"HostSNIRegexp\" matcher, %q is not a valid pattern: %w", template, err)
		}

		regexp, err := regexp.Compile(preparedPattern)
		if err != nil {
			return err
		}

		regexps = append(regexps, regexp)
	}

	tree.matcher = func(meta ConnData) bool {
		for _, regexp := range regexps {
			if regexp.MatchString(meta.serverName) {
				return true
			}
		}

		return false
	}

	return nil
}

// preparePattern builds a regexp pattern from the initial user defined expression.
// This function reuses the code dedicated to host matching of the newRouteRegexp func from the gorilla/mux library.
// https://github.com/containous/mux/tree/8ffa4f6d063c1e2b834a73be6a1515cca3992618.
func preparePattern(template string) (string, error) {
	// Check if it is well-formed.
	idxs, errBraces := braceIndices(template)
	if errBraces != nil {
		return "", errBraces
	}

	defaultPattern := "[^.]+"
	pattern := bytes.NewBufferString("")

	// Host SNI matching is case-insensitive
	_, _ = fmt.Fprint(pattern, "(?i)")

	pattern.WriteByte('^')
	var end int
	for i := 0; i < len(idxs); i += 2 {
		// Set all values we are interested in.
		raw := template[end:idxs[i]]
		end = idxs[i+1]
		parts := strings.SplitN(template[idxs[i]+1:end-1], ":", 2)
		name := parts[0]

		patt := defaultPattern
		if len(parts) == 2 {
			patt = parts[1]
		}

		// Name or pattern can't be empty.
		if name == "" || patt == "" {
			return "", fmt.Errorf("mux: missing name or pattern in %q",
				template[idxs[i]:end])
		}

		// Build the regexp pattern.
		_, _ = fmt.Fprintf(pattern, "%s(?P<%s>%s)", regexp.QuoteMeta(raw), varGroupName(i/2), patt)
	}

	// Add the remaining.
	raw := template[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	pattern.WriteByte('$')

	return pattern.String(), nil
}

// varGroupName builds a capturing group name for the indexed variable.
// This function is a copy of varGroupName func from the gorilla/mux library.
// https://github.com/containous/mux/tree/8ffa4f6d063c1e2b834a73be6a1515cca3992618.
func varGroupName(idx int) string {
	return "v" + strconv.Itoa(idx)
}

// braceIndices returns the first level curly brace indices from a string.
// This function is a copy of braceIndices func from the gorilla/mux library.
// https://github.com/containous/mux/tree/8ffa4f6d063c1e2b834a73be6a1515cca3992618.
func braceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := range len(s) {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}
