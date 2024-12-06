package api

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/traefik/traefik/v3/pkg/version"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func (h Handler) supportDump(rw http.ResponseWriter, request *http.Request) {
	staticConfig, err := json.Marshal(h.staticConfig)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	runtimeConfig, err := json.Marshal(h.runtimeConfiguration)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	tVersion, err := json.Marshal(struct {
		Version   string
		Codename  string
		StartDate time.Time
	}{
		Version:   version.Version,
		Codename:  version.Codename,
		StartDate: version.StartDate,
	})
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
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
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := addFile(tw, "static-config.json", staticConfig); err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := addFile(tw, "runtime-config.json", runtimeConfig); err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func addFile(tw *tar.Writer, name string, content []byte) error {
	header := &tar.Header{
		Name:    name,
		Mode:    0600,
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
