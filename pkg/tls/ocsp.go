package tls

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ocsp"
)

const defaultCacheDuration = 24 * time.Hour

type ocspEntry struct {
	leaf       *x509.Certificate
	issuer     *x509.Certificate
	responders []string
	nextUpdate time.Time
	staple     []byte
}

// ocspStapler retrieves staples from OCSP responders and store them in an in-memory cache.
// It also updates the staples on a regular basis and before they expire.
type ocspStapler struct {
	client             *http.Client
	cache              cache.Cache
	forceStapleUpdates chan struct{}
	responderOverrides map[string]string
}

// newOCSPStapler creates a new ocspStapler cache.
func newOCSPStapler(responderOverrides map[string]string) *ocspStapler {
	return &ocspStapler{
		client:             &http.Client{Timeout: 10 * time.Second},
		cache:              *cache.New(defaultCacheDuration, 5*time.Minute),
		forceStapleUpdates: make(chan struct{}, 1),
		responderOverrides: responderOverrides,
	}
}

// Run updates the OCSP staples every hours.
func (o *ocspStapler) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return

	case <-o.forceStapleUpdates:
		o.updateStaples(ctx)

	case <-ticker.C:
		o.updateStaples(ctx)
	}
}

// ForceStapleUpdates triggers staple updates in the background instead of waiting for the Run routine to update them.
func (o *ocspStapler) ForceStapleUpdates() {
	select {
	case o.forceStapleUpdates <- struct{}{}:
	default:
	}
}

// GetStaple retrieves the OCSP staple for the corresponding to the given key (public certificate hash).
func (o *ocspStapler) GetStaple(key string) ([]byte, bool) {
	if item, ok := o.cache.Get(key); ok && item != nil {
		if entry, ok := item.(*ocspEntry); ok {
			return entry.staple, true
		}
	}
	return nil, false
}

// Upsert creates a new entry for the given certificate.
// The ocspStapler will then be responsible from retrieving and updating the corresponding OCSP obtainStaple.
func (o *ocspStapler) Upsert(key string, leaf, issuer *x509.Certificate) error {
	if len(leaf.OCSPServer) == 0 {
		return errors.New("leaf certificate does not contain an OCSP server")
	}

	if item, ok := o.cache.Get(key); ok {
		o.cache.Set(key, item, cache.NoExpiration)
		return nil
	}

	var responders []string
	for _, url := range leaf.OCSPServer {
		if len(o.responderOverrides) > 0 {
			if newURL, ok := o.responderOverrides[url]; ok {
				url = newURL
			}
		}
		responders = append(responders, url)
	}

	o.cache.Set(key, &ocspEntry{
		leaf:       leaf,
		issuer:     issuer,
		responders: responders,
	}, cache.NoExpiration)

	return nil
}

// ResetTTL resets the expiration time for all items having no expiration.
// This allows setting a TTL for certificates that do not exist anymore in the dynamic configuration.
// For certificates that are still provided by the dynamic configuration,
// their expiration time will be unset when calling the Upsert method.
func (o *ocspStapler) ResetTTL() {
	for key, item := range o.cache.Items() {
		if item.Expiration > 0 {
			continue
		}

		o.cache.Set(key, item.Object, defaultCacheDuration)
	}
}

func (o *ocspStapler) updateStaples(ctx context.Context) {
	for _, item := range o.cache.Items() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		entry := item.Object.(*ocspEntry)

		if entry.staple != nil && time.Now().Before(entry.nextUpdate) {
			continue
		}

		if err := o.updateStaple(ctx, entry); err != nil {
			log.Error().Err(err).Msgf("Unable to retieve OCSP staple for: %s", entry.leaf.Subject.CommonName)
			continue
		}
	}
}

// obtainStaple obtains the OCSP stable for the given leaf certificate.
func (o *ocspStapler) updateStaple(ctx context.Context, entry *ocspEntry) error {
	ocspReq, err := ocsp.CreateRequest(entry.leaf, entry.issuer, nil)
	if err != nil {
		return fmt.Errorf("creating OCSP request: %w", err)
	}

	for _, responder := range entry.responders {
		logger := log.With().Str("responder", responder).Logger()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, responder, bytes.NewReader(ocspReq))
		if err != nil {
			return fmt.Errorf("creating OCSP request: %w", err)
		}

		req.Header.Set("Content-Type", "application/ocsp-request")

		res, err := o.client.Do(req)
		if err != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			logger.Debug().Err(err).Msg("Unable to obtain OCSP response")
			continue
		}
		defer res.Body.Close()

		if res.StatusCode/100 != 2 {
			logger.Debug().Msgf("Unable to obtain OCSP response due to status code: %d", res.StatusCode)
			continue
		}

		ocspResBytes, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Debug().Err(err).Msg("Unable to read OCSP response bytes")
			continue
		}

		ocspRes, err := ocsp.ParseResponseForCert(ocspResBytes, entry.leaf, entry.issuer)
		if err != nil {
			logger.Debug().Err(err).Msg("Unable to parse OCSP response")
			continue
		}

		entry.staple = ocspResBytes

		// As per RFC 6960, the nextUpdate field is optional.
		if ocspRes.NextUpdate.IsZero() {
			// NextUpdate is not set, the staple should be updated on the next update.
			entry.nextUpdate = time.Now()
		} else {
			entry.nextUpdate = ocspRes.ThisUpdate.Add(ocspRes.NextUpdate.Sub(ocspRes.ThisUpdate) / 2)
		}

		return nil
	}

	return errors.New("no OCSP staple obtained from any responders")
}
