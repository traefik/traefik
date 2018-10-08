// Implements gettext in pure Go with Plural Forms support.

package gettext

import (
	"fmt"
	"os"
	"path"
)

// Translations holds the translations in the different locales your app
// supports. Use NewTranslations to create an instance.
type Translations struct {
	cache    map[string]Catalog
	root     string
	domain   string
	resolver PathResolver
}

// PathResolver resolves a path to a mo file
type PathResolver func(root string, locale string, domain string) string

// DefaultResolver resolves paths in the standard format of:
// <root>/<locale>/LC_MESSAGES/<domain>.mo
func DefaultResolver(root string, locale string, domain string) string {
	return path.Join(root, locale, "LC_MESSAGES", fmt.Sprintf("%s.mo", domain))
}

// NewTranslations is the main entry point for gogettext. Use this to set up
// the locales for your app.
// root is the root of your locale folder, domain the domain you want to load
// and resolver a function that resolves mo file paths.
// If your structure is <root>/<locale>/LC_MESSAGES/<domain>.mo, you can use
// DefaultResolver.
func NewTranslations(root string, domain string, resolver PathResolver) Translations {
	return Translations{
		root:     root,
		resolver: resolver,
		domain:   domain,
		cache:    map[string]Catalog{},
	}
}

// Preload a list of locales (if they're available). This is useful if you want
// to limit IO to a specific time in your app, for example startup. Subsequent
// calls to Preload or Locale using a locale given here will not do any IO.
func (t Translations) Preload(locales ...string) {
	for _, locale := range locales {
		t.load(locale)
	}
}

func (t Translations) load(locale string) {
	path := t.resolver(t.root, locale, t.domain)
	f, err := os.Open(path)
	if err != nil {
		t.cache[locale] = nullcatalog{}
		return
	}
	defer f.Close()
	catalog, err := ParseMO(f)
	if err != nil {
		t.cache[locale] = nullcatalog{}
		return
	}
	t.cache[locale] = catalog
}

// Locale returns the catalog translations for a given Locale. If the given
// locale is not available, a NullCatalog is returned.
func (t Translations) Locale(locale string) Catalog {
	_, ok := t.cache[locale]
	if !ok {
		t.load(locale)
	}
	return t.cache[locale]
}
