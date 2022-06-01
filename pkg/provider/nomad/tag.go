package nomad

import (
	"strings"
)

func tagsToLabels(tags []string, prefix string) map[string]string {
	labels := make(map[string]string, len(tags))
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			if parts := strings.SplitN(tag, "=", 2); len(parts) == 2 {
				left, right := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
				key := "traefik." + strings.TrimPrefix(left, prefix+".")
				labels[key] = right
			}
		}
	}
	return labels
}
