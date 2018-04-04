package consulcatalog

import (
	"strings"

	"github.com/containous/traefik/provider/label"
)

func tagsToNeutralLabels(tags []string, prefix string) map[string]string {
	var labels map[string]string

	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {

			parts := strings.SplitN(tag, "=", 2)
			if len(parts) == 2 {
				if labels == nil {
					labels = make(map[string]string)
				}

				// replace custom prefix by the generic prefix
				key := label.Prefix + strings.TrimPrefix(parts[0], prefix+".")
				labels[key] = parts[1]
			}
		}
	}

	return labels
}
