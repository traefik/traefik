package consulcatalog

import "strings"

func inArray(needle string, arr []string) bool {
	for _, s := range arr {
		if s == needle {
			return true
		}
	}
	return false
}

// Example:
// needle: 		"traefik.protocol="
// arr: 		{"bar.bas=foo", "traefik.protocol=tcp", "foo.bar"}
// returns: 	"tcp", true
func inArrayPrefix(needle string, arr []string) (string, bool) {
	for _, s := range arr {
		if strings.HasPrefix(s, needle) {
			return s[len(needle):], true
		}
	}
	return "", false
}
