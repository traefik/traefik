package version

import (
	"context"
	"net/url"

	"github.com/containous/traefik/log"
	"github.com/google/go-github/github"
	goversion "github.com/hashicorp/go-version"
)

var (
	// Version holds the current version of traefik.
	Version = "dev"
	// Codename holds the current version codename of traefik.
	Codename = "cheddar" // beta cheese
	// BuildDate holds the build date of traefik.
	BuildDate = "I don't remember exactly"
)

// CheckNewVersion checks if a new version is available
func CheckNewVersion() {
	if Version == "dev" {
		return
	}
	client := github.NewClient(nil)
	updateURL, err := url.Parse("https://update.traefik.io")
	if err != nil {
		log.Warnf("Error checking new version: %s", err)
		return
	}
	client.BaseURL = updateURL
	releases, resp, err := client.Repositories.ListReleases(context.Background(), "containous", "traefik", nil)
	if err != nil {
		log.Warnf("Error checking new version: %s", err)
		return
	}

	if resp.StatusCode != 200 {
		log.Warnf("Error checking new version: status=%s", resp.Status)
		return
	}

	currentVersion, err := goversion.NewVersion(Version)
	if err != nil {
		log.Warnf("Error checking new version: %s", err)
		return
	}

	for _, release := range releases {
		releaseVersion, err := goversion.NewVersion(*release.TagName)
		if err != nil {
			log.Warnf("Error checking new version: %s", err)
			return
		}

		if len(currentVersion.Prerelease()) == 0 && len(releaseVersion.Prerelease()) > 0 {
			continue
		}

		if releaseVersion.GreaterThan(currentVersion) {
			log.Warnf("A new release has been found: %s. Please consider updating.", releaseVersion.String())
			return
		}
	}
}
