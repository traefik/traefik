package encryption

import (
	"crypto/aes"
	"encoding/base64"
	"testing"

	"crypto/cipher"
	"github.com/stretchr/testify/assert"
)

func TestAesEncrypterEncrypt(t *testing.T) {

	aesKey := "RDFXVxTgrrT9IseypJrwDLzk/nTVeTjbjaUR3RVyv94="

	audit := `
	{
		"eventId": "ev123",
		"auditSource": "foo",
		"auditType": "bar",
		"field1": "field1value"
	}
	`
	enc, _ := NewEncrypter(aesKey)
	result, _ := enc.Encrypt([]byte(audit))
	decoded, err := base64.StdEncoding.DecodeString(result)

	if assert.NoError(t, err) {
		assert.Equal(t, len(decoded), len(audit)+aes.BlockSize)
	}

	key, _ := base64.StdEncoding.DecodeString(aesKey)
	block, err := aes.NewCipher(key)
	if assert.NoError(t, err) {
		iv := decoded[:aes.BlockSize]
		decrypted := make([]byte, len(audit))
		cipher.NewCFBDecrypter(block, iv).XORKeyStream(decrypted, decoded[aes.BlockSize:])

		assert.Equal(t, audit, string(decrypted))
	}
}

func TestNoopEncrypterEncrypt(t *testing.T) {
	enc, _ := NewEncrypter("")
	result, _ := enc.Encrypt([]byte("Some old string"))
	assert.Equal(t, "Some old string", result)
}
