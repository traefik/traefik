package tls

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/types"
	"golang.org/x/crypto/ocsp"
)

type OCSP struct {
	certificate       *tls.Certificate
	config            types.OCSPConfig
	issuedCertificate *x509.Certificate
	issuerCertificate *x509.Certificate
	lock              sync.RWMutex

	MustStaple bool
	Response   *ocsp.Response
	Server     []string
}

// Constants for PKIX MustStaple extension.
var (
	tlsFeatureExtensionOID = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 24}
	ocspMustStapleFeature  = []byte{0x30, 0x03, 0x02, 0x01, 0x05}
	mustStapleExtension    = pkix.Extension{
		Id:    tlsFeatureExtensionOID,
		Value: ocspMustStapleFeature,
	}
)

func NewOCSP(config types.OCSPConfig, certificate *tls.Certificate) (*OCSP, error) {
	issuedCertificate, issuerCertificate, err := parseCertificate(certificate)
	if err != nil {
		return nil, err
	}

	mustStaple := false
	for _, ext := range issuedCertificate.ExtraExtensions {
		if ext.Id.Equal(mustStapleExtension.Id) && reflect.DeepEqual(ext.Value, mustStapleExtension.Value) {
			mustStaple = true
		}
	}

	ocspServer := issuedCertificate.OCSPServer
	if len(config.ResponderOverrides) > 0 {
		for i, respURL := range issuedCertificate.OCSPServer {
			if override, ok := config.ResponderOverrides[respURL]; ok {
				ocspServer[i] = override
			}
		}
	}

	return &OCSP{
		certificate:       certificate,
		config:            config,
		issuedCertificate: issuedCertificate,
		issuerCertificate: issuerCertificate,
		lock:              sync.RWMutex{},

		MustStaple: mustStaple,
		Response:   nil,
		Server:     ocspServer,
	}, nil
}

// Staple populates the ocsp response of the certificate if needed and not disabled by configuration.
func (o *OCSP) Staple() error {
	if o.config.DisableStapling {
		return nil
	}

	o.lock.RLock()
	ocspResponse := o.Response
	o.lock.RUnlock()

	if ocspResponse != nil && time.Now().Before(ocspResponse.ThisUpdate.Add(ocspResponse.NextUpdate.Sub(ocspResponse.ThisUpdate)/2)) {
		return nil
	}

	o.lock.Lock()
	defer o.lock.Unlock()

	ocspRespBytes, ocspResp, ocspErr := o.request()
	if ocspErr != nil {
		return fmt.Errorf("no OCSP stapling for %v: %w", o.issuedCertificate.Subject.CommonName, ocspErr)
	}

	log.Debug().Msgf("ocsp response: %v", ocspResp)
	if ocspResp.Status == ocsp.Good {
		if ocspResp.NextUpdate.After(o.issuedCertificate.NotAfter) {
			return fmt.Errorf("invalid: OCSP response for %v valid after certificate expiration (%s)", o.issuedCertificate.Subject.CommonName, o.issuedCertificate.NotAfter.Sub(ocspResp.NextUpdate))
		}

		o.certificate.OCSPStaple = ocspRespBytes
		o.Response = ocspResp
	}

	// @todo check revocation status

	return nil
}

func (o *OCSP) request() ([]byte, *ocsp.Response, error) {
	ocspReq, err := ocsp.CreateRequest(o.issuedCertificate, o.issuerCertificate, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating OCSP request: %w", err)
	}

	if o.Server == nil {
		return nil, nil, errors.New("no OCSP server specified in certificate")
	}

	reader := bytes.NewReader(ocspReq)
	resp, err := http.Post(o.Server[0], "application/ocsp-request", reader)
	if err != nil {
		return nil, nil, fmt.Errorf("making OCSP request: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, nil, fmt.Errorf("response error: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	ocspResBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, nil, fmt.Errorf("reading OCSP response: %w", err)
	}

	ocspRes, err := ocsp.ParseResponse(ocspResBytes, o.issuerCertificate)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing OCSP response: %w", err)
	}

	return ocspResBytes, ocspRes, nil
}

func parseCertificate(certificate *tls.Certificate) (*x509.Certificate, *x509.Certificate, error) {
	leaf, _ := x509.ParseCertificate(certificate.Certificate[0])
	var issuerCertificate *x509.Certificate
	if len(certificate.Certificate) == 1 {
		issuerCertificate = leaf
	} else {
		ic, err := x509.ParseCertificate(certificate.Certificate[1])
		if err != nil {
			return nil, nil, fmt.Errorf("cannot parse issuer certificate for %v: %w", leaf.Subject.CommonName, err)
		}

		issuerCertificate = ic
	}

	return leaf, issuerCertificate, nil
}
