package tls

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

//////////////////////
//
//   CRL snapshot
//
//////////////////////

// crlSnapshot is an immutable structure containing a CRL state.
// it is replaced atomically on CRL refresh,
// enabling readers to use previous version without locks.
type crlSnapshotCommon struct {
	// Crl serial
	number *big.Int

	// Revoked certificates indexed map, used for O(1) search
	revokedSerials map[string]revocationInfo

	// time at wich this crl has been refreshed
	modTime time.Time

	// Next update contains crl nextUpdate data
	nextUpdate time.Time
}

func (s *crlSnapshotCommon) containsSerial(serial *big.Int) bool {
	_, revoked := s.revokedSerials[serial.String()]
	return revoked
}

func (s *crlSnapshotCommon) nextUpdateTime() time.Time {
	return s.nextUpdate
}

func (s *crlSnapshotCommon) serialCount() int {
	return len(s.revokedSerials)
}

func (s *crlSnapshotCommon) crlSerial() *big.Int {
	return s.number
}

type revocationInfo struct {
	revokedAt time.Time
}

// static crl snapshot, ignoring next update in refresh logic
type staticCrlSnapshot struct {
	crlSnapshotCommon
}

func (s *staticCrlSnapshot) needsRefresh(interval time.Duration) bool {
	return false
}

// dynamic crl snaopshot, refreshs according to nextUpdate
type dynamicCrlSnapshot struct {
	crlSnapshotCommon
}

func (s *dynamicCrlSnapshot) needsRefresh(interval time.Duration) bool {
	if time.Since(s.modTime) > interval && interval > 0 {
		// CRL is fresh but updating in case of a new published CRL
		// ignoring if interval is 0 or less
		return true
	}
	if !s.nextUpdate.IsZero() && time.Now().After(s.nextUpdate) {
		return true
	}
	return false
}

// crlSnapshot is an immutable structure containing a CRL state.
// it is replaced atomically on CRL refresh,
// enabling readers to use previous version without locks.
type crlSnapshot interface {
	// Does this CRL contains a certificate serial ?
	//
	// True means certificate is revoked
	containsSerial(serial *big.Int) bool

	// Does this entry need a refresh (client cert validation time)
	needsRefresh(interval time.Duration) bool

	nextUpdateTime() time.Time

	serialCount() int
	crlSerial() *big.Int
}

type crlSnapshotHolder struct {
	snapshot crlSnapshot
}

//////////////////////
//
//   CRL entry
//
//////////////////////

// crlEntry encapsulates current snapshot and a refresh mutex.
type crlEntry struct {
	// snapshotPtr is a lock free crl snaphot.
	// Readers uses atomic.LoadPointer, without taking lock.
	snapshotPtr atomic.Pointer[crlSnapshotHolder]

	// Emitting CA certificate.
	// used to verify CRL is correctly signed.
	// used by loaders.
	issuerPtr atomic.Pointer[x509.Certificate]

	// refreshMu ensures only one goroutine refreshes this entry.
	refreshMu sync.Mutex

	// CRL loader used to refresh this entry
	loader crlLoader

	// lastFailureUnixNano stores the UnixNano timestamp of the last failed reload
	// attempt (0 if none/cleared). Used to implement a short negative cache so that
	// a persistently unreachable distribution point is not re-fetched on every
	// handshake.
	lastFailureUnixNano atomic.Int64
}

// get current snapshot lock free
func (e *crlEntry) snapshot() crlSnapshot {
	h := e.snapshotPtr.Load()
	if h == nil {
		return nil
	}
	return h.snapshot
}

// store a new snapshot lock free
func (e *crlEntry) storeSnapshot(s crlSnapshot) {
	e.snapshotPtr.Store(&crlSnapshotHolder{snapshot: s})
}

// recordFailure marks a reload attempt as failed, starting (or restarting) the
// backoff window.
func (e *crlEntry) recordFailure() {
	e.lastFailureUnixNano.Store(time.Now().UnixNano())
}

// clearFailure resets failure tracking after a successful reload.
func (e *crlEntry) clearFailure() {
	e.lastFailureUnixNano.Store(0)
}

// inBackoff reports whether a new reload attempt should be skipped because a
// previous failure is still within the configured backoff window. When skip is
// true, retryAt indicates when attempts will be allowed again (for diagnostics).
func (e *crlEntry) inBackoff(backoff time.Duration) (skip bool, retryAt time.Time) {
	if backoff <= 0 {
		return false, time.Time{}
	}
	last := e.lastFailureUnixNano.Load()
	if last == 0 {
		return false, time.Time{}
	}
	retryAt = time.Unix(0, last).Add(backoff)
	return time.Now().Before(retryAt), retryAt
}

func (e *crlEntry) issuer() *x509.Certificate {
	return e.issuerPtr.Load()
}

// updateIssuer store the new CRL emitter and tells wether the CRL emitter changed.
// First record (old == nil) does not need a force refresh.
func (e *crlEntry) updateIssuer(issuer *x509.Certificate) (changed bool) {
	old := e.issuerPtr.Swap(issuer)
	return old != nil && !old.Equal(issuer)
}

// Loads a new snapshot non locking
func (e *crlEntry) reload(distributionPoint string) error {

	snap, err := e.loader.Load(e)
	if err != nil {
		e.recordFailure()
		log.Error().
			Err(err).
			Str("distributionPoint", distributionPoint).
			Msg("Could not reload CRL")
		return err
	}
	e.clearFailure()
	e.storeSnapshot(snap)

	if snap.nextUpdateTime().Before(time.Now()) {
		log.Warn().
			Str("distributionPoint", distributionPoint).
			Int("revokedCount", snap.serialCount()).
			Time("nextUpdate", snap.nextUpdateTime()).
			Msg("CRL is outdated")
	}

	log.Info().
		Str("distributionPoint", distributionPoint).
		Int("revokedCount", snap.serialCount()).
		Time("nextUpdate", snap.nextUpdateTime()).
		Msg("CRL loaded")

	return nil
}

//////////////////////
//
//   CRL store
//
//////////////////////

// A group of CRL distribution points values
type CRLStore struct {
	// crls snapshots (optimized for high read throughput)
	// mapped by distribution point
	entries sync.Map

	// Interval between manual refresh of all held entries
	crlReloadInterval atomic.Pointer[time.Duration]

	// crlErrorBackoff bounds how often a distribution point is retried after a
	// resolution/download failure, preventing a persistently unreachable CRL
	// endpoint from being hit on every TLS handshake.
	crlErrorBackoff atomic.Pointer[time.Duration]
}

func NewFileCRLStore(files map[string]string, reloadInterval *time.Duration, errorBackoff *time.Duration) (*CRLStore, error) {
	store := &CRLStore{}
	store.crlReloadInterval.Store(reloadInterval)
	store.crlErrorBackoff.Store(errorBackoff)

	err := store.loadFiles(files)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func NewCRLStore(reloadInterval *time.Duration, errorBackoff *time.Duration) *CRLStore {
	store := &CRLStore{}
	store.crlReloadInterval.Store(reloadInterval)
	store.crlErrorBackoff.Store(errorBackoff)

	return store
}

// getOrCreateEntry returns existing entry for a distribution point (empty if no entry is known).
func (s *CRLStore) getOrCreateEntry(distributionPoint string) *crlEntry {
	entry := &crlEntry{loader: crlNOOPLoader{}}
	// ensure only one entry is created for each distribution point
	actual, _ := s.entries.LoadOrStore(distributionPoint, entry)
	return actual.(*crlEntry)
}

// Get entry if exists, returns (nil,false) if not found
func (s *CRLStore) getEntry(distributionPoint string) (*crlEntry, bool) {
	actual, ok := s.entries.Load(distributionPoint)
	if !ok {
		return nil, false
	}
	return actual.(*crlEntry), true
}

// snapshot provided files
func (s *CRLStore) loadFiles(files map[string]string) error {
	for distributionPoint, path := range files {
		err := s.loadFile(distributionPoint, path)
		if err != nil {
			return err
		}
	}
	return nil
}

// snapshot a CRL file
//
// locks distribution point entry
//
// !!! Should only be used when creating global store
func (s *CRLStore) loadFile(distributionPoint, filePath string) error {
	// get entry if exists
	entry := s.getOrCreateEntry(distributionPoint)
	// lock to avoid concurrent write
	entry.refreshMu.Lock()
	defer entry.refreshMu.Unlock()

	// set loader of this entry to file
	entry.loader = &crlFileLoader{path: filePath}

	// load and store
	return entry.reload(distributionPoint)
}

// Start CRL manual reload watcher
func (s *CRLStore) WatchEntries(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("stopping CRL watch loop")
			return
		case <-time.After(*s.crlReloadInterval.Load()):
			s.reload()
		}
	}
}

// Reload all held entries
func (s *CRLStore) reload() {
	s.entries.Range(func(key, value any) bool {
		entry := value.(*crlEntry)

		// lock to avoid concurrent write
		entry.refreshMu.Lock()
		defer entry.refreshMu.Unlock()

		// ignore error as it is already logged
		_ = entry.reload(key.(string))

		return true
	})
}

//FIXME consider another file watcher mechnism for kube envs ?

// Watch crl files for a store.
//
// CRL entries will be individually synchronised according to fs events changes.
//
// Please note that in some fs driver contexts, file changes will not be detected.
func (s *CRLStore) WatchFiles(ctx context.Context, files map[string]string) (func(ctx context.Context), error) {
	// init file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	for _, path := range files {
		if err := watcher.Add(path); err != nil {
			return nil, fmt.Errorf("watching %q: %w", path, err)
		}
	}

	// watcher
	return func(ctx context.Context) {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					log.Info().Str("file", event.Name).Msg("CRL file changed, reloading")
					// find dp of modified path
					var distributionPoint string
					for dp, path := range files {
						if path == event.Name {
							distributionPoint = dp
						}
					}
					// return if not found
					if distributionPoint == "" {
						continue
					}
					entry := s.getOrCreateEntry(distributionPoint)
					// lock to avoid concurrent write
					entry.refreshMu.Lock()

					// ignore error as it is already logged
					entry.reload(distributionPoint)
					entry.refreshMu.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error().Err(err).Msg("CRL file watcher error")
			}
		}
	}, nil
}

// parseCRL allows DER or PEM format CRL
func parseCRL(raw []byte) (*x509.RevocationList, error) {
	if crl, err := x509.ParseRevocationList(raw); err == nil {
		return crl, nil
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("data is neither valid DER nor PEM")
	}
	return x509.ParseRevocationList(block.Bytes)
}
