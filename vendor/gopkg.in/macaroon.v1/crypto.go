package macaroon

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

func keyedHash(key *[hashLen]byte, text []byte) *[hashLen]byte {
	h := keyedHasher(key)
	h.Write([]byte(text))
	var sum [hashLen]byte
	hashSum(h, &sum)
	return &sum
}

func keyedHasher(key *[hashLen]byte) hash.Hash {
	return hmac.New(sha256.New, key[:])
}

var keyGen = []byte("macaroons-key-generator")

// makeKey derives a fixed length key from a variable
// length key. The keyGen constant is the same
// as that used in libmacaroons.
func makeKey(variableKey []byte) *[keyLen]byte {
	h := hmac.New(sha256.New, keyGen)
	h.Write(variableKey)
	var key [keyLen]byte
	hashSum(h, &key)
	return &key
}

// hashSum calls h.Sum to put the sum into
// the given destination. It also sanity
// checks that the result really is the expected
// size.
func hashSum(h hash.Hash, dest *[hashLen]byte) {
	r := h.Sum(dest[:0])
	if len(r) != len(dest) {
		panic("hash size inconsistency")
	}
}

const (
	keyLen   = 32
	nonceLen = 24
	hashLen  = sha256.Size
)

func newNonce(r io.Reader) (*[nonceLen]byte, error) {
	var nonce [nonceLen]byte
	_, err := r.Read(nonce[:])
	if err != nil {
		return nil, fmt.Errorf("cannot generate random bytes: %v", err)
	}
	return &nonce, nil
}

func encrypt(key *[keyLen]byte, text *[hashLen]byte, r io.Reader) ([]byte, error) {
	nonce, err := newNonce(r)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(nonce)+secretbox.Overhead+len(text))
	out = append(out, nonce[:]...)
	return secretbox.Seal(out, text[:], nonce, key), nil
}

func decrypt(key *[keyLen]byte, ciphertext []byte) (*[hashLen]byte, error) {
	if len(ciphertext) < nonceLen+secretbox.Overhead {
		return nil, fmt.Errorf("message too short")
	}
	var nonce [nonceLen]byte
	copy(nonce[:], ciphertext)
	ciphertext = ciphertext[nonceLen:]
	text, ok := secretbox.Open(nil, ciphertext, &nonce, key)
	if !ok {
		return nil, fmt.Errorf("decryption failure")
	}
	if len(text) != hashLen {
		return nil, fmt.Errorf("decrypted text is wrong length")
	}
	var rtext [hashLen]byte
	copy(rtext[:], text)
	return &rtext, nil
}
