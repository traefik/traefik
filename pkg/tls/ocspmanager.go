package tls

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	neturl "net/url"

	"golang.org/x/crypto/ocsp"
)

var errCertificateOCSPRevoked = errors.New("certificate revoked by OCSP")

type ocspManager struct {
}

func (om *ocspManager) RevocationCheck(cert *x509.Certificate) (error, error) {
	urls := cert.OCSPServer
	if len(urls) == 0 {
		// OCSP not available for this certificate
		return nil, nil
	}

	issuer := certificateIssuer(cert)
	if issuer == nil {
		return nil, errors.New("no certificate issuer found")
	}

	req, err := ocsp.CreateRequest(cert, issuer, nil)
	if err != nil {
		return nil, err
	}

	for _, u := range urls {
		resp, err := om.sendRequest(u, req, cert, issuer)
		if err != nil {
			// This is a soft-fail
			return err, nil
		}

		if resp.Status != ocsp.Good {
			// Revoked
			return nil, errCertificateOCSPRevoked
		}
	}

	return nil, nil
}

// Adapted from https://github.com/cloudflare/cfssl/blob/a538700acfbab2ae160df7c145207f0c3114e5d3/revoke/revoke.go#L273-L315
func (om *ocspManager) sendRequest(url string, req []byte, cert, issuer *x509.Certificate) (*ocsp.Response, error) {
	var resp *http.Response
	var err error
	if len(req) > 256 {
		buf := bytes.NewBuffer(req)
		resp, err = http.Post(url, "application/ocsp-request", buf)
	} else {
		u := url + "/" + neturl.QueryEscape(base64.StdEncoding.EncodeToString(req))
		resp, err = http.Get(u)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	switch {
	case bytes.Equal(body, ocsp.MalformedRequestErrorResponse):
		return nil, errors.New("malformed")
	case bytes.Equal(body, ocsp.InternalErrorErrorResponse):
		return nil, errors.New("internal error")
	case bytes.Equal(body, ocsp.TryLaterErrorResponse):
		return nil, errors.New("try later")
	case bytes.Equal(body, ocsp.SigRequredErrorResponse):
		return nil, errors.New("signature required")
	case bytes.Equal(body, ocsp.UnauthorizedErrorResponse):
		return nil, errors.New("unauthorized")
	}

	return ocsp.ParseResponseForCert(body, cert, issuer)
}

func newOCSPManager() *ocspManager {
	return &ocspManager{}
}
