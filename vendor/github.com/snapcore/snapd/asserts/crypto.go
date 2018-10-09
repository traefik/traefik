// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package asserts

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha256" // be explicit about supporting SHA256
	_ "crypto/sha512" // be explicit about needing SHA512
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/sha3"
)

const (
	maxEncodeLineLength = 76
	v1                  = 0x1
)

var (
	v1Header         = []byte{v1}
	v1FixedTimestamp = time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)
)

func encodeV1(data []byte) []byte {
	buf := new(bytes.Buffer)
	buf.Grow(base64.StdEncoding.EncodedLen(len(data) + 1))
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	enc.Write(v1Header)
	enc.Write(data)
	enc.Close()
	flat := buf.Bytes()
	flatSize := len(flat)

	buf = new(bytes.Buffer)
	buf.Grow(flatSize + flatSize/maxEncodeLineLength + 1)
	off := 0
	for {
		endOff := off + maxEncodeLineLength
		if endOff > flatSize {
			endOff = flatSize
		}
		buf.Write(flat[off:endOff])
		off = endOff
		if off >= flatSize {
			break
		}
		buf.WriteByte('\n')
	}

	return buf.Bytes()
}

type keyEncoder interface {
	keyEncode(w io.Writer) error
}

func encodeKey(key keyEncoder, kind string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := key.keyEncode(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot encode %s: %v", kind, err)
	}
	return encodeV1(buf.Bytes()), nil
}

type openpgpSigner interface {
	sign(content []byte) (*packet.Signature, error)
}

func signContent(content []byte, privateKey PrivateKey) ([]byte, error) {
	signer, ok := privateKey.(openpgpSigner)
	if !ok {
		panic(fmt.Errorf("not an internally supported PrivateKey: %T", privateKey))
	}

	sig, err := signer.sign(content)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = sig.Serialize(buf)
	if err != nil {
		return nil, err
	}

	return encodeV1(buf.Bytes()), nil
}

func decodeV1(b []byte, kind string) (packet.Packet, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("cannot decode %s: no data", kind)
	}
	buf := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
	n, err := base64.StdEncoding.Decode(buf, b)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %v", kind, err)
	}
	if n == 0 {
		return nil, fmt.Errorf("cannot decode %s: base64 without data", kind)
	}
	buf = buf[:n]
	if buf[0] != v1 {
		return nil, fmt.Errorf("unsupported %s format version: %d", kind, buf[0])
	}
	rd := bytes.NewReader(buf[1:])
	pkt, err := packet.Read(rd)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %v", kind, err)
	}
	if rd.Len() != 0 {
		return nil, fmt.Errorf("%s has spurious trailing data", kind)
	}
	return pkt, nil
}

func decodeSignature(signature []byte) (*packet.Signature, error) {
	pkt, err := decodeV1(signature, "signature")
	if err != nil {
		return nil, err
	}
	sig, ok := pkt.(*packet.Signature)
	if !ok {
		return nil, fmt.Errorf("expected signature, got instead: %T", pkt)
	}
	return sig, nil
}

// PublicKey is the public part of a cryptographic private/public key pair.
type PublicKey interface {
	// ID returns the id of the key used for lookup.
	ID() string

	// verify verifies signature is valid for content using the key.
	verify(content []byte, sig *packet.Signature) error

	keyEncoder
}

type openpgpPubKey struct {
	pubKey   *packet.PublicKey
	sha3_384 string
}

func (opgPubKey *openpgpPubKey) ID() string {
	return opgPubKey.sha3_384
}

func (opgPubKey *openpgpPubKey) verify(content []byte, sig *packet.Signature) error {
	h := sig.Hash.New()
	h.Write(content)
	return opgPubKey.pubKey.VerifySignature(h, sig)
}

func (opgPubKey openpgpPubKey) keyEncode(w io.Writer) error {
	return opgPubKey.pubKey.Serialize(w)
}

func newOpenPGPPubKey(intPubKey *packet.PublicKey) *openpgpPubKey {
	h := sha3.New384()
	h.Write(v1Header)
	err := intPubKey.Serialize(h)
	if err != nil {
		panic("internal error: cannot compute public key sha3-384")
	}
	sha3_384, err := EncodeDigest(crypto.SHA3_384, h.Sum(nil))
	if err != nil {
		panic("internal error: cannot compute public key sha3-384")
	}
	return &openpgpPubKey{pubKey: intPubKey, sha3_384: sha3_384}
}

// RSAPublicKey returns a database useable public key out of rsa.PublicKey.
func RSAPublicKey(pubKey *rsa.PublicKey) PublicKey {
	intPubKey := packet.NewRSAPublicKey(v1FixedTimestamp, pubKey)
	return newOpenPGPPubKey(intPubKey)
}

// DecodePublicKey deserializes a public key.
func DecodePublicKey(pubKey []byte) (PublicKey, error) {
	pkt, err := decodeV1(pubKey, "public key")
	if err != nil {
		return nil, err
	}
	pubk, ok := pkt.(*packet.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected public key, got instead: %T", pkt)
	}
	rsaPubKey, ok := pubk.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected RSA public key, got instead: %T", pubk.PublicKey)
	}
	return RSAPublicKey(rsaPubKey), nil
}

// EncodePublicKey serializes a public key, typically for embedding in an assertion.
func EncodePublicKey(pubKey PublicKey) ([]byte, error) {
	return encodeKey(pubKey, "public key")
}

// PrivateKey is a cryptographic private/public key pair.
type PrivateKey interface {
	// PublicKey returns the public part of the pair.
	PublicKey() PublicKey

	keyEncoder
}

type openpgpPrivateKey struct {
	privk *packet.PrivateKey
}

func (opgPrivK openpgpPrivateKey) PublicKey() PublicKey {
	return newOpenPGPPubKey(&opgPrivK.privk.PublicKey)
}

func (opgPrivK openpgpPrivateKey) keyEncode(w io.Writer) error {
	return opgPrivK.privk.Serialize(w)
}

var openpgpConfig = &packet.Config{
	DefaultHash: crypto.SHA512,
}

func (opgPrivK openpgpPrivateKey) sign(content []byte) (*packet.Signature, error) {
	privk := opgPrivK.privk
	sig := new(packet.Signature)
	sig.PubKeyAlgo = privk.PubKeyAlgo
	sig.Hash = openpgpConfig.Hash()
	sig.CreationTime = time.Now()

	h := openpgpConfig.Hash().New()
	h.Write(content)

	err := sig.Sign(h, privk, openpgpConfig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func decodePrivateKey(privKey []byte) (PrivateKey, error) {
	pkt, err := decodeV1(privKey, "private key")
	if err != nil {
		return nil, err
	}
	privk, ok := pkt.(*packet.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("expected private key, got instead: %T", pkt)
	}
	if _, ok := privk.PrivateKey.(*rsa.PrivateKey); !ok {
		return nil, fmt.Errorf("expected RSA private key, got instead: %T", privk.PrivateKey)
	}
	return openpgpPrivateKey{privk}, nil
}

// RSAPrivateKey returns a PrivateKey for database use out of a rsa.PrivateKey.
func RSAPrivateKey(privk *rsa.PrivateKey) PrivateKey {
	intPrivk := packet.NewRSAPrivateKey(v1FixedTimestamp, privk)
	return openpgpPrivateKey{intPrivk}
}

// GenerateKey generates a private/public key pair.
func GenerateKey() (PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	return RSAPrivateKey(priv), nil
}

func encodePrivateKey(privKey PrivateKey) ([]byte, error) {
	return encodeKey(privKey, "private key")
}

// externally held key pairs

type extPGPPrivateKey struct {
	pubKey         PublicKey
	from           string
	pgpFingerprint string
	bitLen         int
	doSign         func(content []byte) ([]byte, error)
}

func newExtPGPPrivateKey(exportedPubKeyStream io.Reader, from string, sign func(content []byte) ([]byte, error)) (*extPGPPrivateKey, error) {
	var pubKey *packet.PublicKey

	rd := packet.NewReader(exportedPubKeyStream)
	for {
		pkt, err := rd.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("cannot read exported public key: %v", err)
		}
		cand, ok := pkt.(*packet.PublicKey)
		if ok {
			if cand.IsSubkey {
				continue
			}
			if pubKey != nil {
				return nil, fmt.Errorf("cannot select exported public key, found many")
			}
			pubKey = cand
		}
	}

	if pubKey == nil {
		return nil, fmt.Errorf("cannot read exported public key, found none (broken export)")

	}

	rsaPubKey, ok := pubKey.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not a RSA key")
	}

	return &extPGPPrivateKey{
		pubKey:         RSAPublicKey(rsaPubKey),
		from:           from,
		pgpFingerprint: fmt.Sprintf("%X", pubKey.Fingerprint),
		bitLen:         rsaPubKey.N.BitLen(),
		doSign:         sign,
	}, nil
}

func (expk *extPGPPrivateKey) fingerprint() string {
	return expk.pgpFingerprint
}

func (expk *extPGPPrivateKey) PublicKey() PublicKey {
	return expk.pubKey
}

func (expk *extPGPPrivateKey) keyEncode(w io.Writer) error {
	return fmt.Errorf("cannot access external private key to encode it")
}

func (expk *extPGPPrivateKey) sign(content []byte) (*packet.Signature, error) {
	if expk.bitLen < 4096 {
		return nil, fmt.Errorf("signing needs at least a 4096 bits key, got %d", expk.bitLen)
	}

	out, err := expk.doSign(content)
	if err != nil {
		return nil, err
	}

	badSig := fmt.Sprintf("bad %s produced signature: ", expk.from)

	sigpkt, err := packet.Read(bytes.NewBuffer(out))
	if err != nil {
		return nil, fmt.Errorf(badSig+"%v", err)
	}

	sig, ok := sigpkt.(*packet.Signature)
	if !ok {
		return nil, fmt.Errorf(badSig+"got %T", sigpkt)
	}

	if sig.Hash != crypto.SHA512 {
		return nil, fmt.Errorf(badSig + "expected SHA512 digest")
	}

	err = expk.pubKey.verify(content, sig)
	if err != nil {
		return nil, fmt.Errorf(badSig+"it does not verify: %v", err)
	}

	return sig, nil
}
