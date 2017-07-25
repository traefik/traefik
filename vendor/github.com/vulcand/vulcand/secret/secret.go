package secret

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

func NewKeyString() (string, error) {
	k, err := newKey()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(k), nil
}

func KeyFromString(key string) (*[keyLength]byte, error) {
	bytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return decodeKey(bytes)
}

type Box struct {
	key *[32]byte
}

type SealedBytes struct {
	Val   []byte
	Nonce []byte
}

func NewBoxFromKeyString(keyS string) (*Box, error) {
	key, err := KeyFromString(keyS)
	if err != nil {
		return nil, err
	}
	return NewBox(key)
}

func NewBox(bytes *[keyLength]byte) (*Box, error) {
	return &Box{key: bytes}, nil
}

func (b *Box) Seal(value []byte) (*SealedBytes, error) {
	var nonce [nonceLength]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, fmt.Errorf("unable to generate random string: %v", err)
	}
	var encrypted []byte
	encrypted = secretbox.Seal(encrypted[:0], value, &nonce, b.key)
	return &SealedBytes{
		Val:   encrypted,
		Nonce: nonce[:],
	}, nil
}

func (b *Box) Open(e *SealedBytes) ([]byte, error) {
	nonce, err := decodeNonce(e.Nonce)
	if err != nil {
		return nil, err
	}
	var decrypted []byte
	var ok bool
	decrypted, ok = secretbox.Open(decrypted[:0], e.Val, nonce, b.key)
	if !ok {
		return nil, fmt.Errorf("unable to decrypt message")
	}
	return decrypted, nil
}

func decodeNonce(bytes []byte) (*[nonceLength]byte, error) {
	if len(bytes) != nonceLength {
		return nil, fmt.Errorf("wrong nonce length: %d", len(bytes))
	}
	var nonceBytes [nonceLength]byte
	copy(nonceBytes[:], bytes)
	return &nonceBytes, nil
}

func decodeKey(bytes []byte) (*[keyLength]byte, error) {
	if len(bytes) != keyLength {
		return nil, fmt.Errorf("wrong key length: %d", len(bytes))
	}
	var keyBytes [keyLength]byte
	copy(keyBytes[:], bytes)
	return &keyBytes, nil
}

func newKey() ([]byte, error) {
	var bytes [keyLength]byte
	_, err := io.ReadFull(rand.Reader, bytes[:])
	if err != nil {
		return nil, fmt.Errorf("unable to generate random string: %v", err)
	}
	return bytes[:], nil
}

const (
	nonceLength = 24
	keyLength   = 32
)
