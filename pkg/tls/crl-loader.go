package tls

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

// CRL loading mechanism
type crlLoader interface {
	// Load a CRL according to Loader provided config.
	//
	// Creates a new to be stored snapshot
	Load(entry *crlEntry) (crlSnapshot, error)
}

//////////////////////
//
//   NO OPeration loader
//
//////////////////////

// CRL NOOP loader (dummy loader)
type crlNOOPLoader struct {
}

// Create a dummy snapshot
func (l crlNOOPLoader) Load(entry *crlEntry) (crlSnapshot, error) {
	return &staticCrlSnapshot{
		crlSnapshotCommon: crlSnapshotCommon{
			number:         big.NewInt(0),
			revokedSerials: map[string]revocationInfo{},
			nextUpdate:     time.Now(),
			modTime:        time.Now(),
		},
	}, nil
}

//////////////////////
//
//   File Loader
//
//////////////////////

// CRL loader for file based CRLs
type crlFileLoader struct {
	path string
}

// Load CRL from file
func (l *crlFileLoader) Load(entry *crlEntry) (crlSnapshot, error) {
	cleanPath := filepath.Clean(l.path)

	resolvedPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("resolving CRL path %q: %w", l.path, err)
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("stating CRL file %q: %w", l.path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("CRL path %q is not a regular file", l.path)
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("reading CRL file %q: %w", l.path, err)
	}

	crl, err := parseCRL(data)
	if err != nil {
		return nil, fmt.Errorf("parsing CRL file %q: %w", l.path, err)
	}

	revokedSerials := make(map[string]revocationInfo, len(crl.RevokedCertificateEntries))
	for _, rev := range crl.RevokedCertificateEntries {
		revokedSerials[rev.SerialNumber.String()] = revocationInfo{revokedAt: rev.RevocationTime}
	}

	return &staticCrlSnapshot{
		crlSnapshotCommon{
			number:         crl.Number,
			revokedSerials: revokedSerials,
			nextUpdate:     crl.NextUpdate,
			modTime:        time.Now(),
		},
	}, err
}

//////////////////////
//
//   HTTP Loader
//
//////////////////////

// CRL loader for http based CRLs
type crlHTTPLoader struct {
	distributionPoint string
	clientHolder      *crlHTTPClientHolder
}

// Load CRL from HTTP endpoint
func (l *crlHTTPLoader) Load(entry *crlEntry) (crlSnapshot, error) {
	issuer := entry.issuer()
	if issuer == nil {
		return nil, fmt.Errorf("no issuer certificate available for CRL %q", l.distributionPoint)
	}
	old := entry.snapshot() // used by anti-rollback control

	raw, err := l.downloadCRL()
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}

	crl, err := parseCRL(raw)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if err := l.verifyCRLSignature(crl, issuer); err != nil {
		return nil, err
	}
	if err := l.verifyCRLFreshness(crl, time.Now()); err != nil {
		return nil, err
	}

	revokedSerials := make(map[string]revocationInfo, len(crl.RevokedCertificateEntries))
	for _, rev := range crl.RevokedCertificateEntries {
		revokedSerials[rev.SerialNumber.String()] = revocationInfo{revokedAt: rev.RevocationTime}
	}

	snap := &dynamicCrlSnapshot{
		crlSnapshotCommon{
			number:         crl.Number,
			revokedSerials: revokedSerials,
			nextUpdate:     crl.NextUpdate,
			modTime:        time.Now(),
		},
	}

	if old != nil {
		if err := l.verifyCRLNumberMonotonic(snap, old); err != nil {
			return nil, err
		}
	}
	return snap, nil
}

// Download CRL as raw bytes from an http resource
func (l *crlHTTPLoader) downloadCRL() ([]byte, error) {
	crlClient := l.clientHolder.get()
	resp, err := crlClient.client.Get(l.distributionPoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	// Reads maxCRLBytes+1 to detect size overflow
	buf := make([]byte, crlClient.maxCRLBytes+1)
	limitedReader := io.LimitReader(resp.Body, crlClient.maxCRLBytes+1)

	n, err := io.ReadFull(limitedReader, buf)
	switch {
	case err == io.ErrUnexpectedEOF || err == io.EOF:
		// read size is inferior to allocated buffer size
		// continue
	case err != nil:
		return nil, err
	}

	if int64(n) > crlClient.maxCRLBytes {
		return nil, fmt.Errorf("CRL size exceeds the maximum allowed size of %d bytes", crlClient.maxCRLBytes)
	}

	return buf[:n], nil
}

// Check if CRL was signed by emitter.
func (l *crlHTTPLoader) verifyCRLSignature(crl *x509.RevocationList, issuer *x509.Certificate) error {
	if err := crl.CheckSignatureFrom(issuer); err != nil {
		return fmt.Errorf("CRL signature verification failed: %w", err)
	}

	// Issuer declared in CRL must be certificate emitter
	if !bytes.Equal(crl.RawIssuer, issuer.RawSubject) {
		return fmt.Errorf("CRL issuer mismatch: got %q, expected %q",
			crl.Issuer.String(), issuer.Subject.String())
	}

	return nil
}

// Check if CRL is temporally valid
func (l *crlHTTPLoader) verifyCRLFreshness(crl *x509.RevocationList, now time.Time) error {
	if now.Before(crl.ThisUpdate) {
		return fmt.Errorf("CRL thisUpdate is in the future: %s", crl.ThisUpdate)
	}
	if !crl.NextUpdate.IsZero() && now.After(crl.NextUpdate) {
		return fmt.Errorf("CRL has expired: nextUpdate was %s", crl.NextUpdate)
	}
	return nil
}

// Check if newCRL number exists, is not an older serial than actual snapshot
func (l *crlHTTPLoader) verifyCRLNumberMonotonic(newCRL, oldCRL crlSnapshot) error {
	if oldCRL == nil || oldCRL.crlSerial() == nil || newCRL.crlSerial() == nil {
		return nil // no reference or CA does not use serial for CRLs
	}
	if newCRL.crlSerial().Cmp(oldCRL.crlSerial()) < 0 {
		return fmt.Errorf("CRL rollback detected: new number %s < cached number %s",
			newCRL.crlSerial(), oldCRL.crlSerial())
	}
	return nil
}

//////////////////////
//
//   Client holder
//
//////////////////////

// crlHTTPClientHolder allows the effective *http.Client used by a CRLStore's HTTP based
// loaders to be swapped atomically (e.g. when a TLSOptions' configured Timeout changes),
// without requiring already cached crlEntry instances to be touched, locked, or their
// snapshots discarded.
type crlHTTPClientHolder struct {
	client atomic.Pointer[crlHTTPClient]
}

// wrapper for http client and fetch related configs
type crlHTTPClient struct {
	client      http.Client
	maxCRLBytes int64
}

func newCRLHTTPClientHolder(transport *http.Transport, timeout time.Duration, maxCRLBytes int64) *crlHTTPClientHolder {
	h := &crlHTTPClientHolder{}
	h.set(transport, timeout, maxCRLBytes)
	return h
}

// set atomically replaces the client used for subsequent requests. In-flight requests
// using the previous client are unaffected.
func (h *crlHTTPClientHolder) set(transport *http.Transport, timeout time.Duration, maxCRLBytes int64) {
	h.client.Store(&crlHTTPClient{
		client: http.Client{
			Transport:     transport,
			Timeout:       timeout,
			CheckRedirect: h.checkRedirect,
		},
		maxCRLBytes: maxCRLBytes,
	})
}

func (h *crlHTTPClientHolder) checkRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func (h *crlHTTPClientHolder) get() *crlHTTPClient {
	return h.client.Load()
}
