package tls

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"

	"github.com/cloudflare/circl/hpke"
	"golang.org/x/crypto/cryptobyte"
)

// sha256PrivateKeyLength is the required private key length for SHA-256 based ECDH.
const sha256PrivateKeyLength = 32

func UnmarshalECHKey(data []byte) (*tls.EncryptedClientHelloKey, error) {
	var k tls.EncryptedClientHelloKey
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}

		switch block.Type {
		case "PRIVATE KEY":
			k.PrivateKey = block.Bytes
		case "ECHCONFIG":
			k.Config = block.Bytes[2:] // Skip the first two bytes (length prefix)
		default:
			return nil, fmt.Errorf("unknown PEM block %s", block.Type)
		}

		data = rest
	}

	if len(k.Config) == 0 || len(k.PrivateKey) == 0 {
		return nil, fmt.Errorf("missing ECH configuration or private key in PEM file")
	}

	// go ecdh now only supports SHA-256 (32-byte private key)
	if len(k.PrivateKey) < sha256PrivateKeyLength {
		return nil, fmt.Errorf("invalid private key length: expected at least %d bytes, got %d bytes", sha256PrivateKeyLength, len(k.PrivateKey))
	} else if len(k.PrivateKey) > sha256PrivateKeyLength {
		k.PrivateKey = k.PrivateKey[len(k.PrivateKey)-sha256PrivateKeyLength:]
	}

	k.SendAsRetry = true

	return &k, nil
}

func MarshalECHKey(k *tls.EncryptedClientHelloKey) ([]byte, error) {
	if len(k.Config) == 0 || len(k.PrivateKey) == 0 {
		return nil, fmt.Errorf("missing ECH configuration or private key")
	}
	lengthPrefix := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthPrefix, uint16(len(k.Config)))
	configBytes := append(lengthPrefix, k.Config...)
	var pemData []byte
	pemData = append(pemData, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: k.PrivateKey})...)
	pemData = append(pemData, pem.EncodeToMemory(&pem.Block{Type: "ECHCONFIG", Bytes: configBytes})...)

	return pemData, nil
}

type echCipher struct {
	KDFID  uint16
	AEADID uint16
}

type echExtension struct {
	Type uint16
	Data []byte
}

type echConfig struct {
	raw []byte

	Version uint16
	Length  uint16

	ConfigID             uint8
	KemID                uint16
	PublicKey            []byte
	SymmetricCipherSuite []echCipher

	MaxNameLength uint8
	PublicName    []byte
	Extensions    []echExtension
}

func NewECHKey(publicName string) (*tls.EncryptedClientHelloKey, error) {
	publicKey, privateKey, err := hpke.KEM_X25519_HKDF_SHA256.Scheme().GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	publicKeyBytes, err := publicKey.MarshalBinary()
	if err != nil {
		return nil, err
	}
	privateKeyBytes, err := privateKey.MarshalBinary()
	if err != nil {
		return nil, err
	}

	config := echConfig{
		Version:   0xfe0d, // ECH version 0xfe0d
		Length:    0x0000,
		ConfigID:  uint8(rand.Uint()),
		KemID:     uint16(hpke.KEM_X25519_HKDF_SHA256),
		PublicKey: publicKeyBytes,
		SymmetricCipherSuite: []echCipher{
			{KDFID: uint16(hpke.KDF_HKDF_SHA256), AEADID: uint16(hpke.AEAD_AES256GCM)},
		},
		MaxNameLength: 32,
		PublicName:    []byte(publicName),
		Extensions:    nil,
	}
	if len(config.PublicName) > int(config.MaxNameLength) {
		return nil, fmt.Errorf("public name exceeds maximum length of %d bytes", config.MaxNameLength)
	}

	var b cryptobyte.Builder
	b.AddUint16(config.Version) // Version
	b.AddUint16LengthPrefixed(func(b *cryptobyte.Builder) {
		b.AddUint8(config.ConfigID)
		b.AddUint16(config.KemID)
		b.AddUint16(uint16(len(config.PublicKey)))
		b.AddBytes(config.PublicKey)
		b.AddUint16LengthPrefixed(func(c *cryptobyte.Builder) {
			for _, cipher := range config.SymmetricCipherSuite {
				c.AddUint16(cipher.KDFID)
				c.AddUint16(cipher.AEADID)
			}
		})
		b.AddUint8(config.MaxNameLength)
		b.AddUint8(uint8(len(config.PublicName)))
		b.AddBytes(config.PublicName)
		b.AddUint16LengthPrefixed(func(c *cryptobyte.Builder) {
			for _, ext := range config.Extensions {
				c.AddUint16(ext.Type)
				c.AddUint16(uint16(len(ext.Data)))
				c.AddBytes(ext.Data)
			}
		})
	})
	configBytes, err := b.Bytes()
	if err != nil {
		return nil, err
	}

	return &tls.EncryptedClientHelloKey{
		Config:      configBytes,
		PrivateKey:  privateKeyBytes,
		SendAsRetry: true,
	}, nil
}

type ECHRequestConfig[T []byte | string] struct {
	URL      string `description:"The URL to request." json:"u" export:"true"`
	Host     string `description:"The host/sni to request." json:"h" export:"true"`
	ECH      T      `description:"Base64-encoded ECH public configuration list for client use." json:"ech" export:"true"`
	Insecure bool   `description:"If true, skip TLS verification (for testing purposes)." json:"k" export:"true"`
}

// RequestWithECH sends a GET request to a server using the provided ECH configuration.
func RequestWithECH[T []byte | string](c ECHRequestConfig[T]) (body []byte, err error) {
	// Decode the ECH configuration from base64 if it's a string, otherwise use it directly.
	var ech []byte
	if s, ok := any(c.ECH).(string); ok {
		ech, err = base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}
	} else {
		ech = []byte(c.ECH)
	}

	requestURL, err := url.Parse(c.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if c.Host == "" {
		c.Host = requestURL.Hostname()
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:                     c.Host,
				EncryptedClientHelloConfigList: ech,
				MinVersion:                     tls.VersionTLS13,
				InsecureSkipVerify:             c.Insecure,
			},
		},
	}

	req := &http.Request{
		Method: "GET",
		URL:    requestURL,
		Host:   c.Host,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func ECHConfigToConfigList(echConfig []byte) ([]byte, error) {
	var b cryptobyte.Builder
	b.AddUint16LengthPrefixed(func(child *cryptobyte.Builder) {
		child.AddBytes(echConfig)
	})
	return b.Bytes()
}
