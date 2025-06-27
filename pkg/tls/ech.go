package tls

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"github.com/cloudflare/circl/hpke"
	"golang.org/x/crypto/cryptobyte"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
)

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
		return nil, fmt.Errorf("lack of ECH configuration or private key in PEM file")
	}

	// go ecdh now only supports SHA-256 (32-byte private key)
	if len(k.PrivateKey) < 32 {
		return nil, fmt.Errorf("invalid private key length: expected at least 32 bytes, got %d bytes", len(k.PrivateKey))
	} else if len(k.PrivateKey) > 32 {
		k.PrivateKey = k.PrivateKey[len(k.PrivateKey)-32:]
	}

	k.SendAsRetry = true

	return &k, nil
}

func MarshalECHKey(k *tls.EncryptedClientHelloKey) ([]byte, error) {
	if len(k.Config) == 0 || len(k.PrivateKey) == 0 {
		return nil, fmt.Errorf("lack of ECH configuration or private key")
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

func MarshalEncryptedClientHelloConfigList(configs []tls.EncryptedClientHelloKey) ([]byte, error) {
	builder := cryptobyte.NewBuilder(nil)
	builder.AddUint16LengthPrefixed(func(builder *cryptobyte.Builder) {
		for _, c := range configs {
			builder.AddBytes(c.Config)
		}
	})
	return builder.Bytes()
}

// startECHServer starts a TLS server that supports Encrypted Client Hello (ECH).
func startECHServer(bind string, cert tls.Certificate, echKey tls.EncryptedClientHelloKey) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, ECH-enabled TLS server!")
	})

	server := &http.Server{
		Addr:    bind,
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS13,
			EncryptedClientHelloKeys: []tls.EncryptedClientHelloKey{echKey},
		},
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// RequestWithECH sends request to a server using the provided ECH configuration.
func RequestWithECH(rawURL string, host string, echKey tls.EncryptedClientHelloKey) []byte {
	configList, err := MarshalEncryptedClientHelloConfigList([]tls.EncryptedClientHelloKey{echKey})
	if err != nil {
		log.Fatalf("marshal ECH config list error: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:                     host,
				EncryptedClientHelloConfigList: configList,
				MinVersion:                     tls.VersionTLS13,
				InsecureSkipVerify:             true, // For testing purposes, skip TLS verification
			},
		},
	}

	requestURL, _ := url.Parse(rawURL)
	req := &http.Request{
		Method: "GET",
		URL:    requestURL,
		Host:   host,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request error: %v", err)
	}
	defer resp.Body.Close()

	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	fmt.Printf("server response: %s\n", body[:n])
	fmt.Printf("Status code: %d\n", resp.StatusCode)
	fmt.Printf("Response header: %v\n", resp.Header)

	return body[:n]
}
