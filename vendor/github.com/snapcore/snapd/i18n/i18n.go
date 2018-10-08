// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package i18n

//go:generate update-pot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ojii/gettext.go"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/osutil"
)

// TEXTDOMAIN is the message domain used by snappy; see dgettext(3)
// for more information.
var (
	TEXTDOMAIN   = "snappy"
	locale       gettext.Catalog
	translations gettext.Translations
)

func init() {
	bindTextDomain(TEXTDOMAIN, "/usr/share/locale")
	setLocale("")
}

func langpackResolver(baseRoot string, locale string, domain string) string {
	// first check for the real locale (e.g. de_DE)
	// then try to simplify the locale (e.g. de_DE -> de)
	locales := []string{locale, strings.SplitN(locale, "_", 2)[0]}
	for _, locale := range locales {
		r := filepath.Join(locale, "LC_MESSAGES", fmt.Sprintf("%s.mo", domain))

		// look into the core snaps first for translations,
		// then the main system
		candidateDirs := []string{
			filepath.Join(dirs.SnapMountDir, "/core/current/", baseRoot),
			baseRoot,
		}
		for _, root := range candidateDirs {
			// ubuntu uses /usr/lib/locale-langpack and patches the glibc gettext
			// implementation
			langpack := filepath.Join(root, "..", "locale-langpack", r)
			if osutil.FileExists(langpack) {
				return langpack
			}

			regular := filepath.Join(root, r)
			if osutil.FileExists(regular) {
				return regular
			}
		}
	}

	return ""
}

func bindTextDomain(domain, dir string) {
	translations = gettext.NewTranslations(dir, domain, langpackResolver)
}

func setLocale(loc string) {
	if loc == "" {
		loc = os.Getenv("LC_MESSAGES")
		if loc == "" {
			loc = os.Getenv("LANG")
		}
	}
	// de_DE.UTF-8, de_DE@euro all need to get simplified
	loc = strings.Split(loc, "@")[0]
	loc = strings.Split(loc, ".")[0]

	locale = translations.Locale(loc)
}

// G is the shorthand for Gettext
func G(msgid string) string {
	return locale.Gettext(msgid)
}

// https://www.gnu.org/software/gettext/manual/html_node/Plural-forms.html
// (search for 1000)
func ngn(d int) uint32 {
	const max = 1000000
	if d < 0 {
		d = -d
	}
	if d > max {
		return uint32((d % max) + max)
	}
	return uint32(d)
}

// NG is the shorthand for NGettext
func NG(msgid string, msgidPlural string, n int) string {
	return locale.NGettext(msgid, msgidPlural, ngn(n))
}
