package acme

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/lego"
	"github.com/rs/zerolog/log"
)

// getCertificateRenewDurations returns renew durations calculated from the given certificatesDuration in hours.
func getCertificateRenewDurations(certificatesDuration int) (time.Duration, time.Duration) {
	switch {
	case certificatesDuration >= 365*24:
		return 4 * 30 * 24 * time.Hour, 7 * 24 * time.Hour
	case certificatesDuration >= 3*30*24:
		return 30 * 24 * time.Hour, 24 * time.Hour
	case certificatesDuration >= 30*24:
		return 10 * 24 * time.Hour, 12 * time.Hour
	case certificatesDuration >= 6*24:
		return 2 * 24 * time.Hour, 2 * time.Hour
	case certificatesDuration >= 24:
		return 6 * time.Hour, 10 * time.Minute
	default:
		return 20 * time.Minute, time.Minute
	}
}

type certRenewalInfo struct {
	cs       *CertAndStore
	x509Cert *x509.Certificate

	shouldRenew            bool
	renewalID              string
	timeToWaitForNextCheck time.Duration
}

func (crt *certRenewalInfo) isARI() bool {
	return crt.renewalID != ""
}

func certLifetimeHours(crt *x509.Certificate) int {
	return int(math.Round(crt.NotAfter.Sub(crt.NotBefore).Hours()))
}

func shouldRenewBasedOnTime(crt *x509.Certificate, renewalPeriod time.Duration) bool {
	if crt == nil {
		return true
	}
	return crt.NotAfter.Before(time.Now().UTC().Add(renewalPeriod))
}

func (p *Provider) checkARIRenewal(ctx context.Context, x509Cert *x509.Certificate, renewInterval time.Duration) (bool, string, time.Duration, error) {
	if time.Now().After(x509Cert.NotAfter) {
		certID, err := api.MakeARICertID(x509Cert)
		if err != nil {
			return false, "", 0, err
		}
		return true, certID, renewInterval, nil
	}

	client, err := p.getClient()
	if err != nil {
		return false, "", 0, err
	}

	info, err := client.Certificate.GetRenewalInfo(ctx, x509Cert)
	if err != nil {
		return false, "", 0, err
	}

	if info.ExplanationURL != "" {
		log.Ctx(ctx).Info().Msgf("Your certificate authority has provided the following explanation for a renewal: %s", info.ExplanationURL)
	}

	renewAt := info.ShouldRenewAt(time.Now().UTC(), renewInterval)
	if renewAt == nil {
		retryAfter := renewInterval
		if info.RetryAfter > 0 {
			retryAfter = info.RetryAfter
		}
		return false, "", retryAfter, nil
	}

	if renewAt.After(time.Now()) {
		return false, "", renewAt.Sub(time.Now()), nil
	}

	certID, err := api.MakeARICertID(x509Cert)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Could not compute ARI renewal ID, renewing without ARI")
		return true, "", 0, nil
	}

	retryAfter := renewInterval
	if info.RetryAfter > 0 {
		retryAfter = info.RetryAfter
	}

	return true, certID, retryAfter, nil
}

func (p *Provider) renewCertificateWithARI(ctx context.Context, client *lego.Client, crt *certRenewalInfo) (*certificate.Resource, error) {
	logger := log.Ctx(ctx)

	logger.Info().Msgf("Renewing certificate via ARI (replaces %s): %+v", crt.renewalID, crt.cs.Domain)
	domains := certcrypto.ExtractDomains(crt.x509Cert)
	privateKey, err := certcrypto.ParsePEMPrivateKey(crt.cs.Key)
	if err != nil {
		return nil, fmt.Errorf("parsing private key for ARI renewal: %w", err)
	}
	request := certificate.ObtainRequest{
		Domains:        domains,
		Bundle:         true,
		EmailAddresses: p.EmailAddresses,
		Profile:        p.Profile,
		PreferredChain: p.PreferredChain,
		ReplacesCertID: crt.renewalID,
		PrivateKey:     privateKey,
	}
	return client.Certificate.Obtain(ctx, request)
}

func (p *Provider) renewCertificateLegacy(ctx context.Context, client *lego.Client, crt *certRenewalInfo) (*certificate.Resource, error) {
	logger := log.Ctx(ctx)

	logger.Info().Msgf("Renewing ACME certificate: %+v", crt.cs.Domain)

	res := certificate.Resource{
		ID:          crt.cs.Domain.Main,
		Domains:     crt.cs.Domain.ToStrArray(),
		PrivateKey:  crt.cs.Key,
		Certificate: crt.cs.Certificate.Certificate,
	}

	opts := &certificate.RenewOptions{
		Bundle:         true,
		EmailAddresses: p.EmailAddresses,
		Profile:        p.Profile,
		PreferredChain: p.PreferredChain,
	}

	renewedCert, err := client.Certificate.Renew(ctx, res, opts)
	if err != nil {
		return nil, err
	}

	if len(renewedCert.Certificate) == 0 || len(renewedCert.PrivateKey) == 0 {
		return nil, fmt.Errorf("renewed certificate for %v is empty", crt.cs.Domain.ToStrArray())
	}

	return renewedCert, nil
}

func (p *Provider) getRenewalInformation(ctx context.Context, cs *CertAndStore, renewInterval time.Duration) certRenewalInfo {
	ret := certRenewalInfo{cs: cs}
	crt, err := getX509Certificate(ctx, &cs.Certificate)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf("could not parse cert for domain %+v, will renew", cs.Domain)
	}
	ret.x509Cert = crt

	if !p.DisableARI && crt != nil {
		shouldRenew, replacesCertID, retryAfter, err := p.checkARIRenewal(ctx, crt, renewInterval)
		if err != nil {
			if errors.Is(err, api.ErrNoARI) {
				log.Ctx(ctx).Warn().Msg("ACME server does not support ARI, falling back to time-based")
			} else {
				log.Ctx(ctx).Warn().Err(err).Msg("ARI check failed, falling back to time-based")
			}
		} else if shouldRenew {
			ret.shouldRenew = true
			ret.renewalID = replacesCertID
			ret.timeToWaitForNextCheck = retryAfter
			return ret
		} else {
			ret.timeToWaitForNextCheck = retryAfter
			return ret
		}
	}

	if crt == nil {
		ret.shouldRenew = true
		_, timeToWait := getCertificateRenewDurations(p.CertificatesDuration)
		ret.timeToWaitForNextCheck = timeToWait
		return ret
	}

	renewalPeriod, renewalInterval := getCertificateRenewDurations(certLifetimeHours(crt))
	ret.timeToWaitForNextCheck = renewalInterval
	if err != nil || shouldRenewBasedOnTime(crt, renewalPeriod) {
		ret.shouldRenew = true
	}

	return ret
}

func getNextCheckTime(certs []certRenewalInfo, fallback time.Duration) time.Duration {
	var waits []time.Duration
	for _, curr := range certs {
		if curr.timeToWaitForNextCheck > 0 {
			waits = append(waits, curr.timeToWaitForNextCheck)
		}
	}

	if len(waits) == 0 {
		return fallback
	}

	return slices.Min(waits)
}

func (p *Provider) renewCertificates(ctx context.Context, renewInterval time.Duration) time.Duration {
	logger := log.Ctx(ctx)

	logger.Info().Msg("Testing certificate renew...")

	p.certificatesMu.RLock()
	certWorklist := make([]*CertAndStore, len(p.certificates))
	copy(certWorklist, p.certificates)
	p.certificatesMu.RUnlock()

	renewalInfos := make([]certRenewalInfo, len(certWorklist))
	for i, cert := range certWorklist {
		renewalInfos[i] = p.getRenewalInformation(ctx, cert, renewInterval)
	}

	for i := range renewalInfos {
		task := &renewalInfos[i]
		if !task.shouldRenew {
			continue
		}

		client, err := p.getClient()
		if err != nil {
			logger.Info().Err(err).Msgf("Error renewing ACME certificate: %+v", task.cs.Domain)
			continue
		}

		var renewedCert *certificate.Resource
		if task.isARI() {
			renewedCert, err = p.renewCertificateWithARI(ctx, client, task)
		} else {
			renewedCert, err = p.renewCertificateLegacy(ctx, client, task)
		}
		if err != nil {
			logger.Error().Err(err).Msgf("Error renewing ACME certificate: %v", task.cs.Domain)
			continue
		}

		if len(renewedCert.Certificate) == 0 || len(renewedCert.PrivateKey) == 0 {
			logger.Error().Msgf("domains %v renew certificate with no value: %v", task.cs.Domain.ToStrArray(), task.cs)
			continue
		}

		err = p.addCertificateForDomain(task.cs.Domain, renewedCert, task.cs.Store)
		if err != nil {
			logger.Error().Err(err).Msg("Error adding certificate for domain")
			continue
		}

		logger.Info().Msgf("Renewed ACME certificate: %+v", task.cs.Domain)
	}

	return getNextCheckTime(renewalInfos, renewInterval)
}
