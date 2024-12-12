package api

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/redactor"
	"github.com/traefik/traefik/v3/pkg/version"
)

func (h Handler) getSupportDump(rw http.ResponseWriter, req *http.Request) {
	logger := log.Ctx(req.Context())

	staticConfig, err := redactor.Anonymize(h.staticConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to anonymize and marshal static configuration")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	runtimeConfig, err := json.Marshal(h.runtimeConfiguration)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to marshal runtime configuration")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	tVersion, err := json.Marshal(struct {
		Version   string    `json:"version"`
		Codename  string    `json:"codename"`
		StartDate time.Time `json:"startDate"`
	}{
		Version:   version.Version,
		Codename:  version.Codename,
		StartDate: version.StartDate,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Unable to marshal version")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/gzip")
	rw.Header().Set("Content-Disposition", "attachment; filename=support-dump.tar.gz")

	// Create gzip writer.
	gw := gzip.NewWriter(rw)
	defer gw.Close()

	// Create tar writer.
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add configuration files to the archive.
	if err := addFile(tw, "version.json", tVersion); err != nil {
		logger.Error().Err(err).Msg("Unable to archive version file")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := addFile(tw, "static-config.json", []byte(staticConfig)); err != nil {
		logger.Error().Err(err).Msg("Unable to archive static configuration")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := addFile(tw, "runtime-config.json", runtimeConfig); err != nil {
		logger.Error().Err(err).Msg("Unable to archive runtime configuration")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func addFile(tw *tar.Writer, name string, content []byte) error {
	header := &tar.Header{
		Name:    name,
		Mode:    0o600,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("writing tar header: %w", err)
	}

	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("writing tar content: %w", err)
	}

	return nil
}
