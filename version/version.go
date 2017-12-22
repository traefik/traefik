package version

import (
	"context"
	"net/http"
	"net/url"

	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/google/go-github/github"
	goversion "github.com/hashicorp/go-version"
	"github.com/unrolled/render"
)

var (
	// Version holds the current version of traefik.
	Version = "dev"
	// Codename holds the current version codename of traefik.
	Codename = "cheddar" // beta cheese
	// BuildDate holds the build date of traefik.
	BuildDate = "I don't remember exactly"
)

// Handler expose version routes
type Handler struct{}

var (
	templatesRenderer = render.New(render.Options{
		Directory: "nowhere",
	})
)

// AddRoutes add version routes on a router
func (v Handler) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/api/version").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			v := struct {
				Version  string
				Codename string
			}{
				Version:  Version,
				Codename: Codename,
			}
			templatesRenderer.JSON(response, http.StatusOK, v)
		})
}

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

	if resp.StatusCode != http.StatusOK {
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
