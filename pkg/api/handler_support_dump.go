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

func (h Handler) supportDump(rw http.ResponseWriter, request *http.Request) {
	anonStatic, err := redactor.Anonymize(h.staticConfig)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Msg("anonymizing static configuration")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	staticConfig, err := json.Marshal(anonStatic)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Msg("marshaling static configuration")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	runtimeConfig, err := json.Marshal(h.runtimeConfiguration)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Msg("marshaling runtime configuration")
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
		log.Ctx(request.Context()).Error().Err(err).Msg("marshaling version")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/gzip")
	rw.Header().Set("Content-Disposition", "attachment; filename=support-dump.tar.gz")

	// Create gzip writer
	gw := gzip.NewWriter(rw)
	defer gw.Close()

	// Create tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add configuration files to the archive
	if err := addFile(tw, "version.json", tVersion); err != nil {
		log.Ctx(request.Context()).Error().Err(err).Msg("adding version file")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := addFile(tw, "static-config.json", staticConfig); err != nil {
		log.Ctx(request.Context()).Error().Err(err).Msg("adding static configuration file")
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := addFile(tw, "runtime-config.json", runtimeConfig); err != nil {
		log.Ctx(request.Context()).Error().Err(err).Msg("adding runtime configuration file")
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
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("failed to write tar content: %w", err)
	}

	return nil
}
