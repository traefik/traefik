package acme

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/rs/zerolog/log"
)

// getCertificateRenewDurations returns renew durations calculated from the given certificatesDuration in hours.
// The first (RenewPeriod) is the period before the end of the certificate duration, during which the certificate should be renewed.
// The second (RenewInterval) is the interval between renew attempts.
func getCertificateRenewDurations(certificatesDuration int) (time.Duration, time.Duration) {
	switch {
	case certificatesDuration >= 365*24: // >= 1 year
		return 4 * 30 * 24 * time.Hour, 7 * 24 * time.Hour // 4 month, 1 week
	case certificatesDuration >= 3*30*24: // >= 90 days
		return 30 * 24 * time.Hour, 24 * time.Hour // 30 days, 1 day
	case certificatesDuration >= 30*24: // >= 30 days
		return 10 * 24 * time.Hour, 12 * time.Hour // 10 days, 12 hours
	case certificatesDuration >= 6*24: // >= 6 days
		return 2 * 24 * time.Hour, 2 * time.Hour // 2 days, 2 hours
	case certificatesDuration >= 24: // >= 1 days
		return 6 * time.Hour, 10 * time.Minute // 6 hours, 10 minutes
	default:
		return 20 * time.Minute, time.Minute
	}
}

func shouldRenewBasedOnTime(crt *x509.Certificate, renewalPeriod time.Duration) bool {
	if crt == nil {
		return true
	}
	return crt.NotAfter.Before(time.Now().UTC().Add(renewalPeriod))
}

type certRenewalInfo struct {
	cs       *CertAndStore
	x509Cert *x509.Certificate

	shouldRenew            bool
	renewalID              string
	timeToWaitForNextCheck time.Duration
}

func (crt *certRenewalInfo) IsARI() bool {
	return crt.renewalID != ""
}

func certLifetimeHours(crt *x509.Certificate) int {
	start := crt.NotBefore
	end := crt.NotAfter

	return int(math.Round(end.Sub(start).Hours()))
}

// checkARIRenewal queries the ACME ARI endpoint and returns the following:
// - a boolean indicating if the certificate should be renewed
// - a string with the "ReplacesCertID" for the renewal order if a renewal should take place
// - a duration for the next check
// - error if something went wrong
func (p *Provider) checkARIRenewal(ctx context.Context, x509Cert *x509.Certificate, renewInterval time.Duration) (bool, string, time.Duration, error) {
	// per RFC9773 4.3, we should not check ARI if certificate is expired
	if time.Now().After(x509Cert.NotAfter) {
		certID, err := certificate.MakeARICertID(x509Cert)
		if err != nil {
			return false, "", 0, err
		}
		return true, certID, renewInterval, nil
	}

	client, err := p.getClient()
	if err != nil {
		return false, "", 0, err
	}

	info, err := client.Certificate.GetRenewalInfo(certificate.RenewalInfoRequest{
		Cert: x509Cert,
	})
	if err != nil {
		// this may return ErrNoARI in the case that ARI is not supported, caller should check for this
		return false, "", 0, err
	}

	if info.ExplanationURL != "" {
		// per RFC9773 section 4.2, if the URL is present, we should "show it to the operator"...logging is likely good
		// enough
		log.Ctx(ctx).Info().Msgf("Your certificate authority has provided the following explanation for a renewal: %s", info.ExplanationURL)
	}

	renewAt := info.ShouldRenewAt(time.Now().UTC(), renewInterval)
	if renewAt == nil {
		return false, "", info.RetryAfter, nil
	}

	// the ACME server is asking us to renew at some point before we'd normally wake up again, so set our sleep duration
	// to the renewal point
	if renewAt.After(time.Now()) {
		return false, "", renewAt.Sub(time.Now()), nil
	}

	certID, err := certificate.MakeARICertID(x509Cert)
	if err != nil {
		// in this case, we had an error computing the replaces ID. we should still renew, however, so indicate that to
		// the caller
		log.Ctx(ctx).Warn().Err(err).Msg("Could not compute ARI renewal ID, renewing without ARI")
		return true, "", 0, nil
	}

	return true, certID, info.RetryAfter, nil
}

func (p *Provider) renewCertificateWithARI(ctx context.Context, client *lego.Client, crt *certRenewalInfo) (*certificate.Resource, error) {
	logger := log.Ctx(ctx)

	logger.Info().Msgf("Renewing certificate via ARI (replaces %s): %+v", crt.renewalID, crt.cs.Domain)
	domains := certcrypto.ExtractDomains(crt.x509Cert)
	request := certificate.ObtainRequest{
		Domains:        domains,
		Bundle:         true,
		EmailAddresses: p.EmailAddresses,
		Profile:        p.Profile,
		PreferredChain: p.PreferredChain,
		ReplacesCertID: crt.renewalID,
	}
	renewedCert, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}

	return renewedCert, nil
}

func (p *Provider) renewCertificateLegacy(ctx context.Context, client *lego.Client, crt *certRenewalInfo) (*certificate.Resource, error) {
	logger := log.Ctx(ctx)

	logger.Info().Msgf("Renewing ACME certificate: %+v", crt.cs.Domain)

	res := certificate.Resource{
		Domain:      crt.cs.Domain.Main,
		PrivateKey:  crt.cs.Key,
		Certificate: crt.cs.Certificate.Certificate,
	}

	opts := &certificate.RenewOptions{
		Bundle:         true,
		EmailAddresses: p.EmailAddresses,
		Profile:        p.Profile,
		PreferredChain: p.PreferredChain,
	}

	renewedCert, err := client.Certificate.RenewWithOptions(res, opts)
	if err != nil {
		return nil, err
	}

	if len(renewedCert.Certificate) == 0 || len(renewedCert.PrivateKey) == 0 {
		return nil, fmt.Errorf("renewed certificate for %v is empty", crt.cs.Domain.ToStrArray())
	}

	return renewedCert, nil
}

func (p *Provider) getRenewalInformation(ctx context.Context, cs *CertAndStore, renewInterval time.Duration) certRenewalInfo {
	ret := certRenewalInfo{
		cs: cs,
	}
	crt, err := getX509Certificate(ctx, &cs.Certificate)
	if err != nil {
		log.Warn().Ctx(ctx).Err(err).Msgf("could not parse cert for domain %+v, will renew", cs.Domain)
	}
	ret.x509Cert = crt

	if !p.DisableARI && crt != nil {
		shouldRenew, replacesCertID, retryAfter, err := p.checkARIRenewal(ctx, crt, renewInterval)
		if err != nil {
			if errors.Is(err, api.ErrNoARI) {
				log.Warn().Ctx(ctx).Msg("ACME server does not support ARI, falling back to time-based")
			} else {
				log.Warn().Ctx(ctx).Err(err).Msgf("ARI check failed, falling back to time-based")
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

	// if we got here, either ARI is not supported, ARI is disabled, or ARI had an error doing the check. fall back to
	// some time-based logic
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

// getNextCheckTime does the final compute on how long we should sleep before the next check. If we have certificates that
// specify a wait time, use the shortest one. If none specify a valid wait time, just use the fallback.
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

// renewCertificates checks all certificates for renewal (using ARI if applicable). We return a time.Duration that tells
// the main goroutine when to check for renewal next given the last renewInterval. This will likely be the `checkAfter`
// time we get back from ARI, otherwise it will be the old renewInterval.
func (p *Provider) renewCertificates(ctx context.Context, renewInterval time.Duration) time.Duration {
	logger := log.Ctx(ctx)

	logger.Info().Msg("Testing certificate renew...")

	p.certificatesMu.RLock()
	certWorklist := make([]*CertAndStore, len(p.certificates))
	copy(certWorklist, p.certificates)
	p.certificatesMu.RUnlock()

	// we go through every certificate we know and build some renewal information about it. this renewal info will include
	// the decoded certificate, whether the cert needs renewal, and when it should be checked again
	renewalInfos := make([]certRenewalInfo, len(certWorklist))
	for i, cert := range certWorklist {
		renewalInfos[i] = p.getRenewalInformation(ctx, cert, renewInterval)
		logger.Debug().
			Str("domain", cert.Domain.Main).
			Strs("sans", cert.Domain.SANs).
			Bool("should renew", renewalInfos[i].shouldRenew).
			Float64("time_to_wait_mins", renewalInfos[i].timeToWaitForNextCheck.Minutes()).
			Msg("Built renewal info")
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
		if task.IsARI() {
			renewedCert, err = p.renewCertificateWithARI(ctx, client, task)
			if err != nil {
				logger.Info().Err(err).Msgf("Error renewing ACME certificate with ARI: %+v", task.cs.Domain)
				continue
			}
		} else {
			renewedCert, err = p.renewCertificateLegacy(ctx, client, task)
			if err != nil {
				logger.Info().Err(err).Msgf("Error renewing ACME certificate: %+v", task.cs.Domain)
				continue
			}
		}

		// certificate has been renewed, send off the configuration change here,
		err = p.addCertificateForDomain(task.cs.Domain, renewedCert, task.cs.Store)
		if err != nil {
			logger.Error().Err(err).Msg("Error adding certificate for domain")
			continue
		}

		// if a renewed certificate demands a shorter check interval than what we already have, honor it if it was
		// a non-ARI one. if It _WAS_ an ARI-issued certificate, we should check renewal info NOW per RFC 9773 4.3
		renewedX509, err := getX509CertificateFromBytes(renewedCert.Certificate, renewedCert.PrivateKey)
		if err != nil {
			logger.Error().Err(err).Msg("Could not parse renewed certificate after renewal")
			continue
		}

		logger.Info().
			Str("domain", task.cs.Domain.Main).
			Strs("sans", task.cs.Domain.SANs).
			Time("new_expiration", renewedX509.NotAfter).
			Msg("renewed certificate")

		task.x509Cert = renewedX509
		if task.IsARI() {
			renewalShouldRenew, _, renewalNextCheck, err := p.checkARIRenewal(ctx, renewedX509, renewInterval)
			if err != nil {
				logger.Error().Err(err).Msg("Error checking ARI information on renewed certificate")
				continue
			}
			if renewalShouldRenew {
				task.timeToWaitForNextCheck = 1 * time.Second // force this to run very soon
				logger.Warn().
					Str("domain", task.cs.Domain.Main).
					Strs("sans", task.cs.Domain.SANs).
					Msg("ARI says we should renew immediately ... for some reason")
			} else {
				task.timeToWaitForNextCheck = renewalNextCheck
				logger.Debug().
					Str("domain", task.cs.Domain.Main).
					Strs("sans", task.cs.Domain.SANs).
					Float64("new_time_to_check_min", task.timeToWaitForNextCheck.Minutes()).
					Msg("Re-checked ARI")
			}
		} else {
			// if we're not ARI-issued, check again in a given interval based on our lookup table
			_, interval := getCertificateRenewDurations(certLifetimeHours(renewedX509))
			task.timeToWaitForNextCheck = interval
		}
	}

	return getNextCheckTime(renewalInfos, renewInterval)
}
