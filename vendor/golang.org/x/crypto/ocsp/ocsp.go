// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ocsp parses OCSP responses as specified in RFC 2560. OCSP responses
// are signed messages attesting to the validity of a certificate for a small
// period of time. This is used to manage revocation for X.509 certificates.
package ocsp // import "golang.org/x/crypto/ocsp"

import (
	"crypto"
	_ "crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"time"
)

var idPKIXOCSPBasic = asn1.ObjectIdentifier([]int{1, 3, 6, 1, 5, 5, 7, 48, 1, 1})

// These are internal structures that reflect the ASN.1 structure of an OCSP
// response. See RFC 2560, section 4.2.

const (
	ocspSuccess       = 0
	ocspMalformed     = 1
	ocspInternalError = 2
	ocspTryLater      = 3
	ocspSigRequired   = 4
	ocspUnauthorized  = 5
)

type certID struct {
	HashAlgorithm pkix.AlgorithmIdentifier
	NameHash      []byte
	IssuerKeyHash []byte
	SerialNumber  *big.Int
}

type responseASN1 struct {
	Status   asn1.Enumerated
	Response responseBytes `asn1:"explicit,tag:0"`
}

type responseBytes struct {
	ResponseType asn1.ObjectIdentifier
	Response     []byte
}

type basicResponse struct {
	TBSResponseData    responseData
	SignatureAlgorithm pkix.AlgorithmIdentifier
	Signature          asn1.BitString
	Certificates       []asn1.RawValue `asn1:"explicit,tag:0,optional"`
}

type responseData struct {
	Raw           asn1.RawContent
	Version       int              `asn1:"optional,default:1,explicit,tag:0"`
	RequestorName pkix.RDNSequence `asn1:"optional,explicit,tag:1"`
	KeyHash       []byte           `asn1:"optional,explicit,tag:2"`
	ProducedAt    time.Time
	Responses     []singleResponse
}

type singleResponse struct {
	CertID     certID
	Good       asn1.Flag   `asn1:"explicit,tag:0,optional"`
	Revoked    revokedInfo `asn1:"explicit,tag:1,optional"`
	Unknown    asn1.Flag   `asn1:"explicit,tag:2,optional"`
	ThisUpdate time.Time
	NextUpdate time.Time `asn1:"explicit,tag:0,optional"`
}

type revokedInfo struct {
	RevocationTime time.Time
	Reason         int `asn1:"explicit,tag:0,optional"`
}

var (
	oidSignatureMD2WithRSA      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 2}
	oidSignatureMD5WithRSA      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 4}
	oidSignatureSHA1WithRSA     = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 5}
	oidSignatureSHA256WithRSA   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}
	oidSignatureSHA384WithRSA   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 12}
	oidSignatureSHA512WithRSA   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 13}
	oidSignatureDSAWithSHA1     = asn1.ObjectIdentifier{1, 2, 840, 10040, 4, 3}
	oidSignatureDSAWithSHA256   = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 4, 3, 2}
	oidSignatureECDSAWithSHA1   = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 1}
	oidSignatureECDSAWithSHA256 = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}
	oidSignatureECDSAWithSHA384 = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 3}
	oidSignatureECDSAWithSHA512 = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 4}
)

// TODO(agl): this is taken from crypto/x509 and so should probably be exported
// from crypto/x509 or crypto/x509/pkix.
func getSignatureAlgorithmFromOID(oid asn1.ObjectIdentifier) x509.SignatureAlgorithm {
	switch {
	case oid.Equal(oidSignatureMD2WithRSA):
		return x509.MD2WithRSA
	case oid.Equal(oidSignatureMD5WithRSA):
		return x509.MD5WithRSA
	case oid.Equal(oidSignatureSHA1WithRSA):
		return x509.SHA1WithRSA
	case oid.Equal(oidSignatureSHA256WithRSA):
		return x509.SHA256WithRSA
	case oid.Equal(oidSignatureSHA384WithRSA):
		return x509.SHA384WithRSA
	case oid.Equal(oidSignatureSHA512WithRSA):
		return x509.SHA512WithRSA
	case oid.Equal(oidSignatureDSAWithSHA1):
		return x509.DSAWithSHA1
	case oid.Equal(oidSignatureDSAWithSHA256):
		return x509.DSAWithSHA256
	case oid.Equal(oidSignatureECDSAWithSHA1):
		return x509.ECDSAWithSHA1
	case oid.Equal(oidSignatureECDSAWithSHA256):
		return x509.ECDSAWithSHA256
	case oid.Equal(oidSignatureECDSAWithSHA384):
		return x509.ECDSAWithSHA384
	case oid.Equal(oidSignatureECDSAWithSHA512):
		return x509.ECDSAWithSHA512
	}
	return x509.UnknownSignatureAlgorithm
}

// This is the exposed reflection of the internal OCSP structures.

const (
	// Good means that the certificate is valid.
	Good = iota
	// Revoked means that the certificate has been deliberately revoked.
	Revoked = iota
	// Unknown means that the OCSP responder doesn't know about the certificate.
	Unknown = iota
	// ServerFailed means that the OCSP responder failed to process the request.
	ServerFailed = iota
)

// Response represents an OCSP response. See RFC 2560.
type Response struct {
	// Status is one of {Good, Revoked, Unknown, ServerFailed}
	Status                                        int
	SerialNumber                                  *big.Int
	ProducedAt, ThisUpdate, NextUpdate, RevokedAt time.Time
	RevocationReason                              int
	Certificate                                   *x509.Certificate
	// TBSResponseData contains the raw bytes of the signed response. If
	// Certificate is nil then this can be used to verify Signature.
	TBSResponseData    []byte
	Signature          []byte
	SignatureAlgorithm x509.SignatureAlgorithm
}

// CheckSignatureFrom checks that the signature in resp is a valid signature
// from issuer. This should only be used if resp.Certificate is nil. Otherwise,
// the OCSP response contained an intermediate certificate that created the
// signature. That signature is checked by ParseResponse and only
// resp.Certificate remains to be validated.
func (resp *Response) CheckSignatureFrom(issuer *x509.Certificate) error {
	return issuer.CheckSignature(resp.SignatureAlgorithm, resp.TBSResponseData, resp.Signature)
}

// ParseError results from an invalid OCSP response.
type ParseError string

func (p ParseError) Error() string {
	return string(p)
}

// ParseResponse parses an OCSP response in DER form. It only supports
// responses for a single certificate. If the response contains a certificate
// then the signature over the response is checked. If issuer is not nil then
// it will be used to validate the signature or embedded certificate. Invalid
// signatures or parse failures will result in a ParseError.
func ParseResponse(bytes []byte, issuer *x509.Certificate) (*Response, error) {
	var resp responseASN1
	rest, err := asn1.Unmarshal(bytes, &resp)
	if err != nil {
		return nil, err
	}
	if len(rest) > 0 {
		return nil, ParseError("trailing data in OCSP response")
	}

	ret := new(Response)
	if resp.Status != ocspSuccess {
		ret.Status = ServerFailed
		return ret, nil
	}

	if !resp.Response.ResponseType.Equal(idPKIXOCSPBasic) {
		return nil, ParseError("bad OCSP response type")
	}

	var basicResp basicResponse
	rest, err = asn1.Unmarshal(resp.Response.Response, &basicResp)
	if err != nil {
		return nil, err
	}

	if len(basicResp.Certificates) > 1 {
		return nil, ParseError("OCSP response contains bad number of certificates")
	}

	if len(basicResp.TBSResponseData.Responses) != 1 {
		return nil, ParseError("OCSP response contains bad number of responses")
	}

	ret.TBSResponseData = basicResp.TBSResponseData.Raw
	ret.Signature = basicResp.Signature.RightAlign()
	ret.SignatureAlgorithm = getSignatureAlgorithmFromOID(basicResp.SignatureAlgorithm.Algorithm)

	if len(basicResp.Certificates) > 0 {
		ret.Certificate, err = x509.ParseCertificate(basicResp.Certificates[0].FullBytes)
		if err != nil {
			return nil, err
		}

		if err := ret.CheckSignatureFrom(ret.Certificate); err != nil {
			return nil, ParseError("bad OCSP signature")
		}

		if issuer != nil {
			if err := issuer.CheckSignature(ret.Certificate.SignatureAlgorithm, ret.Certificate.RawTBSCertificate, ret.Certificate.Signature); err != nil {
				return nil, ParseError("bad signature on embedded certificate")
			}
		}
	} else if issuer != nil {
		if err := ret.CheckSignatureFrom(issuer); err != nil {
			return nil, ParseError("bad OCSP signature")
		}
	}

	r := basicResp.TBSResponseData.Responses[0]

	ret.SerialNumber = r.CertID.SerialNumber

	switch {
	case bool(r.Good):
		ret.Status = Good
	case bool(r.Unknown):
		ret.Status = Unknown
	default:
		ret.Status = Revoked
		ret.RevokedAt = r.Revoked.RevocationTime
		ret.RevocationReason = r.Revoked.Reason
	}

	ret.ProducedAt = basicResp.TBSResponseData.ProducedAt
	ret.ThisUpdate = r.ThisUpdate
	ret.NextUpdate = r.NextUpdate

	return ret, nil
}

// https://tools.ietf.org/html/rfc2560#section-4.1.1
type ocspRequest struct {
	TBSRequest tbsRequest
}

type tbsRequest struct {
	Version     int `asn1:"explicit,tag:0,default:0"`
	RequestList []request
}

type request struct {
	Cert certID
}

// RequestOptions contains options for constructing OCSP requests.
type RequestOptions struct {
	// Hash contains the hash function that should be used when
	// constructing the OCSP request. If zero, SHA-1 will be used.
	Hash crypto.Hash
}

func (opts *RequestOptions) hash() crypto.Hash {
	if opts == nil || opts.Hash == 0 {
		// SHA-1 is nearly universally used in OCSP.
		return crypto.SHA1
	}
	return opts.Hash
}

// CreateRequest returns a DER-encoded, OCSP request for the status of cert. If
// opts is nil then sensible defaults are used.
func CreateRequest(cert, issuer *x509.Certificate, opts *RequestOptions) ([]byte, error) {
	hashFunc := opts.hash()

	// OCSP seems to be the only place where these raw hash identifiers are
	// used. I took the following from
	// http://msdn.microsoft.com/en-us/library/ff635603.aspx
	var hashOID asn1.ObjectIdentifier
	switch hashFunc {
	case crypto.SHA1:
		hashOID = asn1.ObjectIdentifier([]int{1, 3, 14, 3, 2, 26})
	case crypto.SHA256:
		hashOID = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 1})
	case crypto.SHA384:
		hashOID = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 2})
	case crypto.SHA512:
		hashOID = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 3})
	default:
		return nil, x509.ErrUnsupportedAlgorithm
	}

	if !hashFunc.Available() {
		return nil, x509.ErrUnsupportedAlgorithm
	}
	h := opts.hash().New()

	var publicKeyInfo struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}
	if _, err := asn1.Unmarshal(issuer.RawSubjectPublicKeyInfo, &publicKeyInfo); err != nil {
		return nil, err
	}

	h.Write(publicKeyInfo.PublicKey.RightAlign())
	issuerKeyHash := h.Sum(nil)

	h.Reset()
	h.Write(issuer.RawSubject)
	issuerNameHash := h.Sum(nil)

	return asn1.Marshal(ocspRequest{
		tbsRequest{
			Version: 0,
			RequestList: []request{
				{
					Cert: certID{
						pkix.AlgorithmIdentifier{
							Algorithm:  hashOID,
							Parameters: asn1.RawValue{Tag: 5 /* ASN.1 NULL */},
						},
						issuerNameHash,
						issuerKeyHash,
						cert.SerialNumber,
					},
				},
			},
		},
	})
}
