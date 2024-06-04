package acme

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/ocsp"
)

type OCSP struct {
	client                *http.Client
	defaultRequestOptions *ocsp.RequestOptions
}

func NewOCSP(defaultRequestOptions *ocsp.RequestOptions) *OCSP {
	client := &OCSP{
		client:                &http.Client{Timeout: 1 * time.Second},
		defaultRequestOptions: &ocsp.RequestOptions{Hash: crypto.SHA1},
	}

	if defaultRequestOptions != nil {
		client.defaultRequestOptions = defaultRequestOptions
	}

	return client
}

func (o OCSP) Call(ctx context.Context, certificates []*x509.Certificate) ([]byte, *ocsp.Response, error) {
	return o.do(ctx, certificates, o.defaultRequestOptions)
}

func (o OCSP) do(ctx context.Context, certificates []*x509.Certificate, opts *ocsp.RequestOptions) ([]byte, *ocsp.Response, error) {
	issuedCert := certificates[0]

	if len(issuedCert.OCSPServer) == 0 {
		return nil, nil, errors.New("no OCSP server specified in cert")
	}

	// FIXME
	fmt.Println(issuedCert.OCSPServer)

	issuerCert, err := o.getIssuerCertificate(certificates)
	if err != nil {
		return nil, nil, fmt.Errorf("issuer certificate: %w", err)
	}

	ocspReq, err := ocsp.CreateRequest(issuedCert, issuerCert, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("create OCSP request: %w", err)
	}

	ocspServer := issuedCert.OCSPServer[0]

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ocspServer, bytes.NewReader(ocspReq))
	if err != nil {
		return nil, nil, fmt.Errorf("create an HTTP request for OCSPServer %s: %w", ocspServer, err)
	}

	req.Header.Add("Content-Type", "application/ocsp-request")
	req.Header.Add("Accept", "application/ocsp-response")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("call to OCSPServer %s: %w", ocspServer, err)
	}
	defer func() { _ = resp.Body.Close() }()

	ocspResBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read OCSP response %s: %w", ocspServer, err)
	}

	ocspRes, err := ocsp.ParseResponse(ocspResBytes, issuerCert)
	if err != nil {
		return nil, nil, fmt.Errorf("parse OCSP response %s: %w", ocspServer, err)
	}

	return ocspResBytes, ocspRes, nil
}

func (o OCSP) getIssuerCertificate(certificates []*x509.Certificate) (*x509.Certificate, error) {
	issuingCertificateURL := certificates[0].IssuingCertificateURL

	if len(certificates) > 1 {
		return certificates[1], nil
	}

	if len(issuingCertificateURL) == 0 {
		return nil, errors.New("no issuing certificate URL")
	}

	// TODO(ldez): build fallback.
	req, err := http.NewRequest(http.MethodGet, issuingCertificateURL[0], http.NoBody)
	if err != nil {
		return nil, err
	}

	// FIXME
	fmt.Println(req)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	// FIXME check status code.
	fmt.Println(resp.StatusCode)

	issuerBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(issuerBytes)
}
