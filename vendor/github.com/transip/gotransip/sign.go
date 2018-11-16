package gotransip

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
)

var (
	asn1Header = []byte{
		0x30, 0x51, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03,
		0x04, 0x02, 0x03, 0x05, 0x00, 0x04, 0x40,
	}
)

func signWithKey(params *soapParams, key []byte) (string, error) {
	// create SHA512 hash of given parameters
	h := sha512.New()
	h.Write([]byte(params.Encode()))

	// prefix ASN1 header to SHA512 hash
	digest := append(asn1Header, h.Sum(nil)...)

	// prepare key struct
	block, _ := pem.Decode(key)
	if block == nil {
		return "", errors.New("could not decode private key")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("could not parse private key: %s", err.Error())
	}

	pkey := parsed.(*rsa.PrivateKey)

	enc, err := rsa.SignPKCS1v15(rand.Reader, pkey, crypto.Hash(0), digest)
	if err != nil {
		return "", fmt.Errorf("could not sign data: %s", err.Error())
	}

	return url.QueryEscape(base64.StdEncoding.EncodeToString(enc)), nil
}
