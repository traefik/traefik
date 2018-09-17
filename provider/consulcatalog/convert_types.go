package consulcatalog

import (
	"crypto/sha1"
	"encoding/base64"
	"sort"
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

func tagsWithPrefix(tags []string, prefix string) []string {
	var selectedTags []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			selectedTags = append(selectedTags, tag)
		}
	}
	return selectedTags
}

func tagsToGroupId(tags []string, prefix string) string {
	selectedTags := tagsWithPrefix(tags, prefix)

	hash := sha1.New()
	sort.Strings(selectedTags)
	for _, tag := range selectedTags {
		hash.Write([]byte(tag))
		hash.Write([]byte{0})
	}

	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}
