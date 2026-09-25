package tls

import (
	"crypto/x509"
	"errors"
	"fmt"
	"slices"
)

// Enforce CRL validation against certificates

type CRLEnforcer interface {
	// Checks a whole validated certificate chain against the associated CRls
	//
	// chain[0] is the leaf certificate, chain[len-1] should be the root CA (wich is self signed and thus cannot have CRL)
	IsChainAllowed(chain []*x509.Certificate) (bool, error)
}

// enforcer that does nothing
type crlEnforcerNOOP struct {
}

// All chains are allowed for NOOP enforcer
func (e *crlEnforcerNOOP) IsChainAllowed(chain []*x509.Certificate) (bool, error) {
	return true, nil
}

//////////////////////
//
//   (Strict/Lax) enforcer
//
//////////////////////

// crlEnforcer implements both the "lax" and "strict" CRL enforcement policies.
// The two policies only differ in how they treat certificates that carry no
// CRL distribution point: lax allows them, strict rejects them. All other
// logic (chain traversal, distribution point handling) is identical and is
// therefore factorized here instead of being duplicated across two structs.
type crlEnforcer struct {
	// strict controls whether a certificate without any CRLDistributionPoints
	// is rejected (true) or allowed (false).
	strict bool

	whitelistEnabled             bool
	AllowedCRLDistributionPoints []string
	// local http based store
	store *CRLStore
	// global store with file based CRLs
	globalStore     *CRLStore
	snaphotProvider crlSnapshotProvider
}

// IsChainAllowed checks each certificate in the chain against its issuer.
// Behaviour on an empty chain mirrors the previous lax/strict split: lax
// allows it, strict rejects it (no certificate to validate against).
func (e *crlEnforcer) IsChainAllowed(chain []*x509.Certificate) (bool, error) {
	if len(chain) == 0 {
		return !e.strict, nil
	}
	for i := 0; i < len(chain)-1; i++ {
		cert := chain[i]
		issuer := chain[i+1] // emitter, validated by TLS handshake

		allowed, err := e.isCertAllowedAgainstIssuer(cert, issuer)
		if err != nil {
			return false, fmt.Errorf("checking cert CN=%q: %w", cert.Subject.CommonName, err)
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// isCertAllowedAgainstIssuer checks whether a single certificate is allowed.
//
// RFC 5280 allows a CA to partition revocation data across several
// distribution points (via the Issuing Distribution Point extension) instead
// of mirroring the same list on each one. To account for this, all reachable
// distribution points are checked: the certificate is treated as revoked if
// *any* successfully verified distribution point reports it as such, and is
// only treated as valid if *at least one* distribution point could be
// verified and none of the verified ones reported a revocation.
func (e *crlEnforcer) isCertAllowedAgainstIssuer(crt, issuer *x509.Certificate) (bool, error) {
	if len(crt.CRLDistributionPoints) == 0 {
		if e.strict {
			return false, errors.New("strict mode requires a CRL distribution point, none present on certificate")
		}
		// Lax: no crl distribution point is allowed
		return true, nil
	}

	var lastErr error
	anyVerified := false

	for _, dp := range crt.CRLDistributionPoints {
		snap, err := e.resolveSnapshot(dp, issuer)
		if err != nil {
			lastErr = err
			continue
		}

		anyVerified = true
		if snap.containsSerial(crt.SerialNumber) {
			// Revoked according to at least one verified distribution point.
			return false, nil
		}
	}

	if !anyVerified {
		return false, fmt.Errorf("all CRL distribution points failed: %w", lastErr)
	}

	// All verified distribution points agree the certificate is not revoked.
	return true, nil
}

// resolveSnapshot fetches the CRL snapshot for a given distribution point,
// first checking the global (file based) store, then falling back to the
// local HTTP based store subject to whitelist checks.
func (e *crlEnforcer) resolveSnapshot(dp string, issuer *x509.Certificate) (snapshot crlSnapshot, err error) {
	if entry, ok := e.globalStore.getEntry(dp); ok {
		return entry.snapshot(), nil
	}

	if e.whitelistEnabled && !slices.Contains(e.AllowedCRLDistributionPoints, dp) {
		return nil, fmt.Errorf("CRL URI %q is not in the allow-list", dp)
	}

	snap, err := e.snaphotProvider.getVerifiedSnapshot(e.store, dp, issuer)
	if err != nil {
		return nil, err
	}
	return snap, nil
}
