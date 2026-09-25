package tls

import (
	"crypto/x509"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSnapshotProvider struct {
	snap crlSnapshot
	err  error
}

func (s *stubSnapshotProvider) getVerifiedSnapshot(store *CRLStore, dp string, issuer *x509.Certificate) (crlSnapshot, error) {
	return s.snap, s.err
}

type multiDPProvider struct {
	byDP map[string]crlSnapshot
	errs map[string]error
}

func (p *multiDPProvider) getVerifiedSnapshot(store *CRLStore, dp string, issuer *x509.Certificate) (crlSnapshot, error) {
	if err, ok := p.errs[dp]; ok {
		return nil, err
	}
	return p.byDP[dp], nil
}

func revokedSnapshot(revoked ...int64) crlSnapshot {
	m := make(map[string]revocationInfo, len(revoked))
	for _, s := range revoked {
		m[big.NewInt(s).String()] = revocationInfo{}
	}
	return &staticCrlSnapshot{crlSnapshotCommon{number: big.NewInt(1), revokedSerials: m}}
}

func TestCrlEnforcerNOOP_AlwaysAllowed(t *testing.T) {
	allowed, err := (&crlEnforcerNOOP{}).IsChainAllowed(nil)
	require.NoError(t, err)
	assert.True(t, allowed)
}

// --- Lax (strict: false) ---

func TestCrlEnforcerLax_NoDistributionPoints_Allowed(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, nil)

	e := &crlEnforcer{strict: false, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{}}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCrlEnforcerLax_ValidCert_Allowed(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://dp1"})

	e := &crlEnforcer{strict: false, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{snap: revokedSnapshot(999)}}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCrlEnforcerLax_RevokedCert_Denied(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://dp1"})

	e := &crlEnforcer{strict: false, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{snap: revokedSnapshot(1)}}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestCrlEnforcerLax_GlobalStoreTakesPrecedence(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://dp1"})

	global := &CRLStore{}
	global.getOrCreateEntry("http://dp1").storeSnapshot(revokedSnapshot(1))

	e := &crlEnforcer{
		strict: false, store: &CRLStore{}, globalStore: global,
		snaphotProvider: &stubSnapshotProvider{err: fmt.Errorf("should not be called")},
	}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestCrlEnforcerLax_WhitelistBlocksDisallowedDP(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://not-allowed"})

	e := &crlEnforcer{
		strict: false, whitelistEnabled: true, AllowedCRLDistributionPoints: []string{"http://allowed"},
		store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{snap: revokedSnapshot()},
	}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	assert.False(t, allowed)
	assert.Error(t, err)
}

func TestCrlEnforcerLax_WhitelistAllowsListedDP(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://allowed"})

	e := &crlEnforcer{
		strict: false, whitelistEnabled: true, AllowedCRLDistributionPoints: []string{"http://allowed"},
		store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{snap: revokedSnapshot()},
	}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCrlEnforcerLax_FallsBackToNextDPOnFailure(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://down", "http://up"})

	provider := &multiDPProvider{
		byDP: map[string]crlSnapshot{"http://up": revokedSnapshot()},
		errs: map[string]error{"http://down": fmt.Errorf("unreachable")},
	}

	e := &crlEnforcer{strict: false, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: provider}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCrlEnforcerLax_AllDPsFail_ReturnsError(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://down1", "http://down2"})

	provider := &multiDPProvider{errs: map[string]error{
		"http://down1": fmt.Errorf("unreachable"),
		"http://down2": fmt.Errorf("unreachable"),
	}}

	e := &crlEnforcer{strict: false, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: provider}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	assert.False(t, allowed)
	assert.Error(t, err)
}

// --- Partitioned distribution points ---

func TestCrlEnforcerLax_RevokedOnSecondDP_Denied(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://partition1", "http://partition2"})

	// First DP reports the cert as not revoked, second DP (a different
	// partition) reports it as revoked. The certificate must be denied:
	// a "first success wins" strategy would have incorrectly allowed it.
	provider := &multiDPProvider{
		byDP: map[string]crlSnapshot{
			"http://partition1": revokedSnapshot(999),
			"http://partition2": revokedSnapshot(1),
		},
	}

	e := &crlEnforcer{strict: false, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: provider}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.False(t, allowed)
}

// --- Strict (strict: true) ---

func TestCrlEnforcerStrict_NoDistributionPoints_Denied(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, nil)

	e := &crlEnforcer{strict: true, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{}}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	assert.False(t, allowed)
	assert.Error(t, err)
}

func TestCrlEnforcerStrict_ValidCert_Allowed(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://dp1"})

	e := &crlEnforcer{strict: true, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{snap: revokedSnapshot(999)}}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCrlEnforcerStrict_RevokedCert_Denied(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://dp1"})

	e := &crlEnforcer{strict: true, store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{snap: revokedSnapshot(1)}}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestCrlEnforcerStrict_GlobalStoreTakesPrecedence(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://dp1"})

	global := &CRLStore{}
	global.getOrCreateEntry("http://dp1").storeSnapshot(revokedSnapshot())

	e := &crlEnforcer{
		strict: true, store: &CRLStore{}, globalStore: global,
		snaphotProvider: &stubSnapshotProvider{err: fmt.Errorf("should not be called")},
	}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCrlEnforcerStrict_WhitelistBlocksDisallowedDP(t *testing.T) {
	ca := newTestCA(t)
	leaf := newTestLeaf(t, ca, 1, []string{"http://not-allowed"})

	e := &crlEnforcer{
		strict: true, whitelistEnabled: true, AllowedCRLDistributionPoints: []string{"http://allowed"},
		store: &CRLStore{}, globalStore: &CRLStore{}, snaphotProvider: &stubSnapshotProvider{snap: revokedSnapshot()},
	}

	allowed, err := e.IsChainAllowed([]*x509.Certificate{leaf, ca.cert})
	assert.False(t, allowed)
	assert.Error(t, err)
}
