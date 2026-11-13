package tls

import (
	"crypto/x509"
	"fmt"

	"github.com/rs/zerolog/log"
)

type crlSnapshotProvider interface {
	// get a snapshot from a store for a DP using it's issuer
	getVerifiedSnapshot(store *CRLStore, distributionPoint string, issuer *x509.Certificate) (crlSnapshot, error)
}

// CRL loaded through HTTP should be refreshed by the first request using it,
// other requests will be using stale data
type openSnapshotProvider struct {
	clientHolder *crlHTTPClientHolder
}

func (p *openSnapshotProvider) getVerifiedSnapshot(store *CRLStore, distributionPoint string, issuer *x509.Certificate) (crlSnapshot, error) {
	entry := store.getOrCreateEntry(distributionPoint)
	issuerChanged := entry.updateIssuer(issuer)

	snap := entry.snapshot()

	errorBackoff := *store.crlErrorBackoff.Load()
	reloadInterval := *store.crlReloadInterval.Load()

	// new entry or CRL issuer changed (CRL needs a refresh as it is now signed by the old CA)
	if snap == nil || issuerChanged {
		if skip, retryAt := entry.inBackoff(errorBackoff); skip {
			if snap != nil {
				// Rotation detected but the distribution point is in
				// backoff: keep serving the stale snapshot rather than
				// failing the handshake outright.
				return snap, nil
			}
			return nil, fmt.Errorf("CRL %q unreachable, retry not attempted until %s", distributionPoint, retryAt)
		}
		entry.refreshMu.Lock()
		defer entry.refreshMu.Unlock()
		// check if entry has been loaded by another goroutine
		if current := entry.snapshot(); current != snap {
			return current, nil
		}
		if skip, retryAt := entry.inBackoff(errorBackoff); skip {
			return nil, fmt.Errorf("CRL %q unreachable, retry not attempted until %s", distributionPoint, retryAt)
		}
		// set loader to HTTP only if not already configured as such
		if _, ok := entry.loader.(*crlHTTPLoader); !ok {
			entry.loader = &crlHTTPLoader{
				distributionPoint: distributionPoint,
				clientHolder:      p.clientHolder,
			}
		}
		if err := entry.reload(distributionPoint); err != nil {
			log.Warn().
				Str("distributionPoint", distributionPoint).
				Msgf("could not load CRL %s, %v", distributionPoint, err)
		}
		snap = entry.snapshot()

	} else if snap.needsRefresh(reloadInterval) {
		// existing entry
		if skip, _ := entry.inBackoff(errorBackoff); skip {
			// Backoff active: serve the stale snapshot instead of retrying.
			return snap, nil
		}
		if !entry.refreshMu.TryLock() {
			// if couldn't lock, use stale entry
			return snap, nil
		}
		// lock and reload
		defer entry.refreshMu.Unlock()
		snap = entry.snapshot()
		// check if reloaded by another goroutine
		if !snap.needsRefresh(reloadInterval) {
			return snap, nil
		}
		if skip, _ := entry.inBackoff(errorBackoff); skip {
			return snap, nil
		}

		if err := entry.reload(distributionPoint); err != nil {
			log.Warn().
				Str("distributionPoint", distributionPoint).
				Msgf("could not refresh CRL %s, using stale data: %v", distributionPoint, err)
		}
		snap = entry.snapshot()

	}

	if snap == nil {
		return nil, fmt.Errorf("no valid CRL snapshot available")
	}

	return snap, nil
}

// CRL loaded through HTTP should be refreshed by the first request using it,
// other requests will be locked (expect high latency burst)
type failedCloseSnapshotProvider struct {
	clientHolder *crlHTTPClientHolder
}

func (p *failedCloseSnapshotProvider) getVerifiedSnapshot(store *CRLStore, distributionPoint string, issuer *x509.Certificate) (crlSnapshot, error) {
	entry := store.getOrCreateEntry(distributionPoint)
	issuerRotated := entry.updateIssuer(issuer)

	errorBackoff := *store.crlErrorBackoff.Load()
	reloadInterval := *store.crlReloadInterval.Load()

	// Fast path: snapshot present, fresh, and still validated against the
	// current issuer → lock-free read.
	snap := entry.snapshot()
	if snap != nil && !snap.needsRefresh(reloadInterval) && !issuerRotated {
		return snap, nil
	}

	if skip, retryAt := entry.inBackoff(errorBackoff); skip {
		return nil, fmt.Errorf("CRL %q unreachable, retry not attempted until %s", distributionPoint, retryAt)
	}

	// Absent or expired snaphot, mandatory blocking refresh
	// All goroutines will wait here
	entry.refreshMu.Lock()
	defer entry.refreshMu.Unlock()

	// check if reloaded by another goroutine
	snap = entry.snapshot()
	if snap != nil && !issuerRotated && !snap.needsRefresh(reloadInterval) {
		return snap, nil
	}
	if skip, retryAt := entry.inBackoff(errorBackoff); skip {
		return nil, fmt.Errorf("CRL %q unreachable, retry not attempted until %s", distributionPoint, retryAt)
	}

	if snap == nil {
		// new entry set loader
		entry.loader = &crlHTTPLoader{
			distributionPoint: distributionPoint,
			clientHolder:      p.clientHolder,
		}
	}

	if err := entry.reload(distributionPoint); err != nil {
		// Fail-closed : do not return expired snap
		return nil, fmt.Errorf("%s: %v", "CRL is stale or missing and could not be refreshed", err)
	}

	// updated snapshot
	return entry.snapshot(), nil
}
