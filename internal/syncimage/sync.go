package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-github/v72/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var excludedTags = map[string]struct{}{
	"1.7.9-alpine":    {},
	"v1.0.0-beta.392": {},
	"v1.0.0-beta.404": {},
	"v1.0.0-beta.704": {},
	"v1.0.0-rc1":      {},
	"v1.7.9-alpine":   {},
}

func main() {
	dryRun := flag.Bool("dry-run", true, "only print what would be synced")
	src := flag.String("src", "traefik", "source registry/image")
	dst := flag.String("dst", "ghcr.io/traefik/traefik", "destination registry/image")
	flag.Parse()

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	if err := run(*src, *dst, *dryRun); err != nil {
		log.Fatal().Err(err).Msg("Sync failed")
	}
}

func run(src, dst string, dryRun bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log.Info().Str("src", src).Str("dst", dst).Bool("dry-run", dryRun).Msg("Starting sync")

	// Step 1: bulk-fetch digests from DockerHub REST API.
	// Old tags without a digest in this API are skipped — they are immutable and already synced.
	log.Info().Msg("Fetching digests from DockerHub...")
	srcDigests, err := fetchDockerHubDigests(ctx, src)
	if err != nil {
		return fmt.Errorf("fetching DockerHub digests: %w", err)
	}
	log.Info().Int("tags", len(srcDigests)).Msg("DockerHub digests fetched")

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Step 2: bulk-fetch GHCR digests via GitHub API.
	log.Info().Msg("Fetching digests from GHCR...")
	dstDigests, err := fetchGHCRDigests(ctx)
	if err != nil {
		return fmt.Errorf("fetching GHCR digests: %w", err)
	}
	log.Info().Int("tags", len(dstDigests)).Msg("GHCR digests fetched")

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Step 3: compare and sync.
	var inSync, toSync, synced, skipped, failed int

	for tag, srcD := range srcDigests {
		if ctx.Err() != nil {
			log.Warn().Msg("Interrupted")
			break
		}

		if _, ok := excludedTags[tag]; ok {
			skipped++
			continue
		}

		if srcD == dstDigests[tag] {
			inSync++
			continue
		}

		toSync++

		if dryRun {
			log.Info().Str("tag", tag).Str("src", shorten(srcD)).Str("dst", shorten(dstDigests[tag])).Msg("Would sync (dry-run)")
			continue
		}

		srcRef := fmt.Sprintf("%s:%s", src, tag)
		dstRef := fmt.Sprintf("%s:%s", dst, tag)
		if err := crane.Copy(srcRef, dstRef, crane.WithContext(ctx)); err != nil {
			failed++
			log.Error().Err(err).Str("tag", tag).Msg("Failed to sync")
		} else {
			synced++
			log.Info().Str("tag", tag).Msg("Synced")
		}
	}

	log.Info().
		Int("total", len(srcDigests)).
		Int("in_sync", inSync).
		Int("to_sync", toSync).
		Int("synced", synced).
		Int("skipped", skipped).
		Int("failed", failed).
		Msg("Summary")

	if failed > 0 {
		return fmt.Errorf("%d tags failed to sync", failed)
	}

	return nil
}

// fetchDockerHubDigests fetches tag→digest mappings via the DockerHub REST API.
// Old tags may not have a digest and are skipped.
func fetchDockerHubDigests(ctx context.Context, image string) (map[string]string, error) {
	result := make(map[string]string)
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/library/%s/tags?page_size=100", image)

	for url != "" {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		body, err := httpGetWithRetry(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fetching %s: %w", url, err)
		}

		var resp struct {
			Next    *string `json:"next"`
			Results []struct {
				Name   string `json:"name"`
				Digest string `json:"digest"`
			} `json:"results"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for _, r := range resp.Results {
			if r.Digest != "" {
				result[r.Name] = r.Digest
			}
		}

		log.Info().Int("fetched", len(result)).Msg("DockerHub progress")

		if resp.Next != nil {
			url = *resp.Next
		} else {
			url = ""
		}
	}

	return result, nil
}

// fetchGHCRDigests fetches tag→digest mappings via the GitHub Packages API.
func fetchGHCRDigests(ctx context.Context) (map[string]string, error) {
	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))

	result := make(map[string]string)
	opts := &github.PackageListOptions{
		PackageType: github.Ptr("container"),
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		versions, resp, err := client.Organizations.PackageGetAllVersions(ctx, "traefik", "container", "traefik", opts)
		if err != nil {
			return nil, fmt.Errorf("listing GHCR versions: %w", err)
		}

		for _, v := range versions {
			metadata, ok := v.GetMetadata()
			if !ok || metadata.Container == nil {
				continue
			}

			digest := v.GetName()
			for _, tag := range metadata.Container.Tags {
				result[tag] = digest
			}
		}

		log.Info().Int("fetched", len(result)).Msg("GHCR progress")

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return result, nil
}

func httpGetWithRetry(ctx context.Context, url string) ([]byte, error) {
	var result []byte

	op := func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return backoff.Permanent(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return errors.New("429 Too Many Requests")
		}

		if resp.StatusCode != http.StatusOK {
			return backoff.Permanent(fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)))
		}

		result = body
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0

	err := backoff.Retry(op, backoff.WithContext(backoff.WithMaxRetries(bo, 3), ctx))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func shorten(digest string) string {
	if len(digest) > 19 && strings.HasPrefix(digest, "sha256:") {
		return digest[:19] + "..."
	}
	if digest == "" {
		return "MISSING"
	}
	return digest
}
