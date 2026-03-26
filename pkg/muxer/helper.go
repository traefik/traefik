package muxer

import "strings"

func IsHostEqual(domain string, host string) bool {
	if strings.HasPrefix(host, "*") {
		labels := strings.Split(domain, ".")
		labels[0] = "*"
		return strings.EqualFold(host, strings.Join(labels, "."))
	}

	return strings.EqualFold(domain, host)
}
