package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/containous/traefik/log"
)

type aesEncrypter struct {
	key []byte
}

type noopEncrypter struct{}

// Encrypter applies an encryption process to supplied data and base 64 encodes the result
type Encrypter interface {
	Encrypt(plain []byte) (string, error)
}

// NewEncrypter returns an Encrypter implementation. If an candidate key is nil or zero string then a noop
// implementation will be returned otherwise an AES implementation will be returned.
func NewEncrypter(aesKeyBase64 string) (Encrypter, error) {
	if aesKeyBase64 != "" {
		data, err := base64.StdEncoding.DecodeString(aesKeyBase64)
		var enc aesEncrypter
		if err == nil {
			key := []byte(data)
			enc = aesEncrypter{key}
			log.Info("AES Encryption Enabled")
		}
		return &enc, err
	}

	log.Info("NoOp Encryption Enabled")
	return &noopEncrypter{}, nil
}

// Applies AES streaming encryption to the supplied data and then returns the Base64 representation of that.
func (e *aesEncrypter) Encrypt(plain []byte) (string, error) {

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plain))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plain)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Returns the supplied data as string
func (e *noopEncrypter) Encrypt(plain []byte) (string, error) {
	return string(plain), nil
}
