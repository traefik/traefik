/*-
 * Copyright 2014 Square Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package jose

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"regexp"
	"testing"
)

// Reset random reader to original value
func resetRandReader() {
	randReader = rand.Reader
}

// Build big int from hex-encoded string. Strips whitespace (for testing).
func fromHexInt(base16 string) *big.Int {
	re := regexp.MustCompile(`\s+`)
	val, ok := new(big.Int).SetString(re.ReplaceAllString(base16, ""), 16)
	if !ok {
		panic("Invalid test data")
	}
	return val
}

// Build big int from base64-encoded string. Strips whitespace (for testing).
func fromBase64Int(base64 string) *big.Int {
	re := regexp.MustCompile(`\s+`)
	val, err := base64URLDecode(re.ReplaceAllString(base64, ""))
	if err != nil {
		panic("Invalid test data")
	}
	return new(big.Int).SetBytes(val)
}

// Decode hex-encoded string into byte array. Strips whitespace (for testing).
func fromHexBytes(base16 string) []byte {
	re := regexp.MustCompile(`\s+`)
	val, err := hex.DecodeString(re.ReplaceAllString(base16, ""))
	if err != nil {
		panic("Invalid test data")
	}
	return val
}

// Decode base64-encoded string into byte array. Strips whitespace (for testing).
func fromBase64Bytes(b64 string) []byte {
	re := regexp.MustCompile(`\s+`)
	val, err := base64.StdEncoding.DecodeString(re.ReplaceAllString(b64, ""))
	if err != nil {
		panic("Invalid test data")
	}
	return val
}

// Test vectors below taken from crypto/x509/x509_test.go in the Go std lib.

var pkixPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3VoPN9PKUjKFLMwOge6+
wnDi8sbETGIx2FKXGgqtAKpzmem53kRGEQg8WeqRmp12wgp74TGpkEXsGae7RS1k
enJCnma4fii+noGH7R0qKgHvPrI2Bwa9hzsH8tHxpyM3qrXslOmD45EH9SxIDUBJ
FehNdaPbLP1gFyahKMsdfxFJLUvbUycuZSJ2ZnIgeVxwm4qbSvZInL9Iu4FzuPtg
fINKcbbovy1qq4KvPIrXzhbY3PWDc6btxCf3SE0JdE1MCPThntB62/bLMSQ7xdDR
FF53oIpvxe/SCOymfWq/LW849Ytv3Xwod0+wzAP8STXG4HSELS4UedPYeHJJJYcZ
+QIDAQAB
-----END PUBLIC KEY-----`

var pkcs1PrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBALKZD0nEffqM1ACuak0bijtqE2QrI/KLADv7l3kK3ppMyCuLKoF0
fd7Ai2KW5ToIwzFofvJcS/STa6HA5gQenRUCAwEAAQJBAIq9amn00aS0h/CrjXqu
/ThglAXJmZhOMPVn4eiu7/ROixi9sex436MaVeMqSNf7Ex9a8fRNfWss7Sqd9eWu
RTUCIQDasvGASLqmjeffBNLTXV2A5g4t+kLVCpsEIZAycV5GswIhANEPLmax0ME/
EO+ZJ79TJKN5yiGBRsv5yvx5UiHxajEXAiAhAol5N4EUyq6I9w1rYdhPMGpLfk7A
IU2snfRJ6Nq2CQIgFrPsWRCkV+gOYcajD17rEqmuLrdIRexpg8N1DOSXoJ8CIGlS
tAboUGBxTDq3ZroNism3DaMIbKPyYrAqhKov1h5V
-----END RSA PRIVATE KEY-----`

var ecdsaSHA256p384CertPem = `
-----BEGIN CERTIFICATE-----
MIICSjCCAdECCQDje/no7mXkVzAKBggqhkjOPQQDAjCBjjELMAkGA1UEBhMCVVMx
EzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDU1vdW50YWluIFZpZXcxFDAS
BgNVBAoMC0dvb2dsZSwgSW5jMRcwFQYDVQQDDA53d3cuZ29vZ2xlLmNvbTEjMCEG
CSqGSIb3DQEJARYUZ29sYW5nLWRldkBnbWFpbC5jb20wHhcNMTIwNTIxMDYxMDM0
WhcNMjIwNTE5MDYxMDM0WjCBjjELMAkGA1UEBhMCVVMxEzARBgNVBAgMCkNhbGlm
b3JuaWExFjAUBgNVBAcMDU1vdW50YWluIFZpZXcxFDASBgNVBAoMC0dvb2dsZSwg
SW5jMRcwFQYDVQQDDA53d3cuZ29vZ2xlLmNvbTEjMCEGCSqGSIb3DQEJARYUZ29s
YW5nLWRldkBnbWFpbC5jb20wdjAQBgcqhkjOPQIBBgUrgQQAIgNiAARRuzRNIKRK
jIktEmXanNmrTR/q/FaHXLhWRZ6nHWe26Fw7Rsrbk+VjGy4vfWtNn7xSFKrOu5ze
qxKnmE0h5E480MNgrUiRkaGO2GMJJVmxx20aqkXOk59U8yGA4CghE6MwCgYIKoZI
zj0EAwIDZwAwZAIwBZEN8gvmRmfeP/9C1PRLzODIY4JqWub2PLRT4mv9GU+yw3Gr
PU9A3CHMdEcdw/MEAjBBO1lId8KOCh9UZunsSMfqXiVurpzmhWd6VYZ/32G+M+Mh
3yILeYQzllt/g0rKVRk=
-----END CERTIFICATE-----`

var ecdsaSHA256p384CertDer = fromBase64Bytes(`
MIICSjCCAdECCQDje/no7mXkVzAKBggqhkjOPQQDAjCBjjELMAkGA1UEBhMCVVMx
EzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDU1vdW50YWluIFZpZXcxFDAS
BgNVBAoMC0dvb2dsZSwgSW5jMRcwFQYDVQQDDA53d3cuZ29vZ2xlLmNvbTEjMCEG
CSqGSIb3DQEJARYUZ29sYW5nLWRldkBnbWFpbC5jb20wHhcNMTIwNTIxMDYxMDM0
WhcNMjIwNTE5MDYxMDM0WjCBjjELMAkGA1UEBhMCVVMxEzARBgNVBAgMCkNhbGlm
b3JuaWExFjAUBgNVBAcMDU1vdW50YWluIFZpZXcxFDASBgNVBAoMC0dvb2dsZSwg
SW5jMRcwFQYDVQQDDA53d3cuZ29vZ2xlLmNvbTEjMCEGCSqGSIb3DQEJARYUZ29s
YW5nLWRldkBnbWFpbC5jb20wdjAQBgcqhkjOPQIBBgUrgQQAIgNiAARRuzRNIKRK
jIktEmXanNmrTR/q/FaHXLhWRZ6nHWe26Fw7Rsrbk+VjGy4vfWtNn7xSFKrOu5ze
qxKnmE0h5E480MNgrUiRkaGO2GMJJVmxx20aqkXOk59U8yGA4CghE6MwCgYIKoZI
zj0EAwIDZwAwZAIwBZEN8gvmRmfeP/9C1PRLzODIY4JqWub2PLRT4mv9GU+yw3Gr
PU9A3CHMdEcdw/MEAjBBO1lId8KOCh9UZunsSMfqXiVurpzmhWd6VYZ/32G+M+Mh
3yILeYQzllt/g0rKVRk=`)

var pkcs8ECPrivateKey = `
-----BEGIN PRIVATE KEY-----
MIHtAgEAMBAGByqGSM49AgEGBSuBBAAjBIHVMIHSAgEBBEHqkl65VsjYDQWIHfgv
zQLPa0JZBsaJI16mjiH8k6VA4lgfK/KNldlEsY433X7wIzo43u8OpX7Nv7n8pVRH
15XWK6GBiQOBhgAEAfDuikMI4bWsyse7t8iSCmjt9fneW/qStZuIPuVLo7mSJdud
Cs3J/x9wOnnhLv1u+0atnq5HKKdL4ff3itJPlhmSAQzByKQ5LTvB7d6fn95GJVK/
hNuS5qGBpB7qeMXVFoki0/2RZIOway8/fXjmNYwe4v/XB5LLn4hcTvEUGYcF8M9K
-----END PRIVATE KEY-----`

var ecPrivateKey = `
-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIBv2rdY9mWGD/UgiuXB0LJcUzgaB6TXq/Ra1jrZKBV3IGSacM5QDFu
N8yrywiQaTDEqn1zVcLwrnqoQux3gWN1jxugBwYFK4EEACOhgYkDgYYABAFJgaM/
2a3+gE6Khm/1PYftqNwAzQ21HSLp27q2lTN+GBFho691ARFRkr9UzlQ8gRnhkTbu
yGfASamlHsYlr3Tv+gFc4BY8SU0q8kzpQ0dOHWFk7dfGFmKwhJrSFIIOeRn/LY03
XsVFctNDsGhobS2JguQrxhGx8Ll7vQCakV/PEmCQJA==
-----END EC PRIVATE KEY-----`

var ecPrivateKeyDer = fromBase64Bytes(`
MIHcAgEBBEIBv2rdY9mWGD/UgiuXB0LJcUzgaB6TXq/Ra1jrZKBV3IGSacM5QDFu
N8yrywiQaTDEqn1zVcLwrnqoQux3gWN1jxugBwYFK4EEACOhgYkDgYYABAFJgaM/
2a3+gE6Khm/1PYftqNwAzQ21HSLp27q2lTN+GBFho691ARFRkr9UzlQ8gRnhkTbu
yGfASamlHsYlr3Tv+gFc4BY8SU0q8kzpQ0dOHWFk7dfGFmKwhJrSFIIOeRn/LY03
XsVFctNDsGhobS2JguQrxhGx8Ll7vQCakV/PEmCQJA==`)

var invalidPemKey = `
-----BEGIN PUBLIC KEY-----
MIHcAgEBBEIBv2rdY9mWGD/UgiuXB0LJcUzgaB6TXq/Ra1jrZKBV3IGSacM5QDFu
XsVFctNDsGhobS2JguQrxhGx8Ll7vQCakV/PEmCQJA==
-----END PUBLIC KEY-----`

func TestLoadPublicKey(t *testing.T) {
	pub, err := LoadPublicKey([]byte(pkixPublicKey))
	switch pub.(type) {
	case *rsa.PublicKey:
	default:
		t.Error("failed to parse RSA PKIX public key:", err)
	}

	pub, err = LoadPublicKey([]byte(ecdsaSHA256p384CertPem))
	switch pub.(type) {
	case *ecdsa.PublicKey:
	default:
		t.Error("failed to parse ECDSA X.509 cert:", err)
	}

	pub, err = LoadPublicKey([]byte(ecdsaSHA256p384CertDer))
	switch pub.(type) {
	case *ecdsa.PublicKey:
	default:
		t.Error("failed to parse ECDSA X.509 cert:", err)
	}

	pub, err = LoadPublicKey([]byte("###"))
	if err == nil {
		t.Error("should not parse invalid key")
	}

	pub, err = LoadPublicKey([]byte(invalidPemKey))
	if err == nil {
		t.Error("should not parse invalid key")
	}
}

func TestLoadPrivateKey(t *testing.T) {
	priv, err := LoadPrivateKey([]byte(pkcs1PrivateKey))
	switch priv.(type) {
	case *rsa.PrivateKey:
	default:
		t.Error("failed to parse RSA PKCS1 private key:", err)
	}

	priv, err = LoadPrivateKey([]byte(pkcs8ECPrivateKey))
	if _, ok := priv.(*ecdsa.PrivateKey); !ok {
		t.Error("failed to parse EC PKCS8 private key:", err)
	}

	priv, err = LoadPrivateKey([]byte(ecPrivateKey))
	if _, ok := priv.(*ecdsa.PrivateKey); !ok {
		t.Error("failed to parse EC private key:", err)
	}

	priv, err = LoadPrivateKey([]byte(ecPrivateKeyDer))
	if _, ok := priv.(*ecdsa.PrivateKey); !ok {
		t.Error("failed to parse EC private key:", err)
	}

	priv, err = LoadPrivateKey([]byte("###"))
	if err == nil {
		t.Error("should not parse invalid key")
	}

	priv, err = LoadPrivateKey([]byte(invalidPemKey))
	if err == nil {
		t.Error("should not parse invalid key")
	}
}
