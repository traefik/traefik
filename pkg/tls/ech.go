package tls

import (
	"bytes"
	"crypto/ecdh"
	"crypto/hpke"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"

	"golang.org/x/crypto/cryptobyte"
)

const (
	echConfigVersion       = 0xfe0d
	maxECHPublicNameLength = 253
)

var errUnsupportedMandatoryECHConfigExtension = errors.New("unsupported mandatory ECH configuration extension")

type parsedECHConfig struct {
	version           uint16
	kemID             uint16
	publicKey         []byte
	maximumNameLength uint8
	publicName        []byte
}

// UnmarshalECHKeys parses an RFC 9934 PEM-encoded ECH key, returning one key
// per configuration matching the private key.
func UnmarshalECHKeys(data []byte) ([]tls.EncryptedClientHelloKey, error) {
	privateKeyDER, configList, err := decodeECHPEM(data)
	if err != nil {
		return nil, err
	}

	parsedPrivateKey, err := x509.ParsePKCS8PrivateKey(privateKeyDER)
	if err != nil {
		return nil, fmt.Errorf("parsing ECH private key as PKCS#8: %w", err)
	}

	privateKey, ok := parsedPrivateKey.(*ecdh.PrivateKey)
	if !ok || privateKey.Curve() != ecdh.X25519() {
		return nil, fmt.Errorf("unsupported ECH private key type %T: expected X25519", parsedPrivateKey)
	}
	hpkePrivateKey, err := hpke.NewDHKEMPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("parsing ECH private key as HPKE key: %w", err)
	}
	privateKeyBytes, err := hpkePrivateKey.Bytes()
	if err != nil {
		return nil, fmt.Errorf("serializing ECH private key: %w", err)
	}

	configs, err := parseECHConfigList(configList)
	if err != nil {
		return nil, err
	}

	publicKey := hpkePrivateKey.PublicKey().Bytes()
	var keys []tls.EncryptedClientHelloKey
	for _, config := range configs {
		parsed, err := parseECHConfig(config)
		if err != nil {
			if errors.Is(err, errUnsupportedMandatoryECHConfigExtension) {
				continue
			}

			return nil, err
		}
		if parsed.version != echConfigVersion || parsed.kemID != hpkePrivateKey.KEM().ID() || !bytes.Equal(parsed.publicKey, publicKey) {
			continue
		}

		keys = append(keys, tls.EncryptedClientHelloKey{
			Config:      config,
			PrivateKey:  privateKeyBytes,
			SendAsRetry: true,
		})
	}
	if len(keys) == 0 {
		return nil, errors.New("no ECH configuration matches the private key")
	}

	return keys, nil
}

func decodeECHPEM(data []byte) ([]byte, []byte, error) {
	var sawBlock bool
	var privateKey, configList []byte
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		sawBlock = true

		switch block.Type {
		case "PRIVATE KEY":
			if privateKey != nil {
				return nil, nil, errors.New("multiple ECH private keys in PEM file")
			}
			privateKey = block.Bytes
		case "ECHCONFIG":
			// Content after the ECHConfigList block is ignored (RFC 9934, Section 3).
			configList = block.Bytes
			data = nil
			continue
		default:
			return nil, nil, fmt.Errorf("unknown PEM block %s", block.Type)
		}

		data = rest
	}

	if !sawBlock {
		return nil, nil, errors.New("no PEM block in ECH key data: a value that is not an existing file path is treated as inline content")
	}
	if len(privateKey) == 0 {
		return nil, nil, errors.New("missing ECH private key in PEM file: the PRIVATE KEY block must precede the ECHCONFIG block")
	}
	if len(configList) == 0 {
		return nil, nil, errors.New("missing ECH configuration in PEM file")
	}

	return privateKey, configList, nil
}

// MarshalECHKey encodes an X25519 ECH key in the RFC 9934 PEM format.
func MarshalECHKey(key *tls.EncryptedClientHelloKey) ([]byte, error) {
	if key == nil || len(key.Config) == 0 || len(key.PrivateKey) == 0 {
		return nil, errors.New("missing ECH configuration or private key")
	}

	kem := hpke.DHKEM(ecdh.X25519())
	hpkePrivateKey, err := kem.NewPrivateKey(key.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parsing X25519 HPKE private key: %w", err)
	}
	privateKeyBytes, err := hpkePrivateKey.Bytes()
	if err != nil {
		return nil, fmt.Errorf("serializing X25519 HPKE private key: %w", err)
	}

	config, err := parseECHConfig(key.Config)
	if err != nil {
		return nil, err
	}
	if config.version != echConfigVersion || config.kemID != kem.ID() {
		return nil, errors.New("ECH configuration does not use X25519")
	}
	if !bytes.Equal(config.publicKey, hpkePrivateKey.PublicKey().Bytes()) {
		return nil, errors.New("ECH configuration does not match the private key")
	}

	privateKey, err := ecdh.X25519().NewPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("converting X25519 HPKE private key to PKCS#8: %w", err)
	}
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("marshaling ECH private key as PKCS#8: %w", err)
	}

	configList, err := ECHConfigToConfigList(key.Config)
	if err != nil {
		return nil, err
	}

	var pemData []byte
	pemData = append(pemData, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER})...)
	pemData = append(pemData, pem.EncodeToMemory(&pem.Block{Type: "ECHCONFIG", Bytes: configList})...)

	return pemData, nil
}

// NewECHKey generates an X25519 ECH key with a single configuration for the
// given public name.
func NewECHKey(publicName string) (*tls.EncryptedClientHelloKey, error) {
	if len(publicName) == 0 {
		return nil, errors.New("public name is empty")
	}
	if !validECHPublicName([]byte(publicName)) {
		return nil, fmt.Errorf("invalid ECH public name %q", publicName)
	}

	kem := hpke.DHKEM(ecdh.X25519())
	privateKey, err := kem.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("generating X25519 HPKE key: %w", err)
	}
	privateKeyBytes, err := privateKey.Bytes()
	if err != nil {
		return nil, fmt.Errorf("serializing X25519 HPKE key: %w", err)
	}

	var configID [1]byte
	if _, err = rand.Read(configID[:]); err != nil {
		return nil, fmt.Errorf("generating ECH configuration ID: %w", err)
	}

	var builder cryptobyte.Builder
	builder.AddUint16(echConfigVersion)
	builder.AddUint16LengthPrefixed(func(builder *cryptobyte.Builder) {
		builder.AddUint8(configID[0])
		builder.AddUint16(kem.ID())
		builder.AddUint16LengthPrefixed(func(builder *cryptobyte.Builder) {
			builder.AddBytes(privateKey.PublicKey().Bytes())
		})
		builder.AddUint16LengthPrefixed(func(builder *cryptobyte.Builder) {
			builder.AddUint16(hpke.HKDFSHA256().ID())
			builder.AddUint16(hpke.AES128GCM().ID())
		})
		// The generator cannot know the longest hidden server name in advance.
		builder.AddUint8(0)
		builder.AddUint8LengthPrefixed(func(builder *cryptobyte.Builder) {
			builder.AddBytes([]byte(publicName))
		})
		builder.AddUint16LengthPrefixed(func(*cryptobyte.Builder) {})
	})
	config, err := builder.Bytes()
	if err != nil {
		return nil, fmt.Errorf("marshaling ECH configuration: %w", err)
	}

	return &tls.EncryptedClientHelloKey{
		Config:      config,
		PrivateKey:  privateKeyBytes,
		SendAsRetry: true,
	}, nil
}

// ECHPublicName returns the public name of the ECH key configuration.
func ECHPublicName(key tls.EncryptedClientHelloKey) (string, bool) {
	parsed, err := parseECHConfig(key.Config)
	if err != nil || parsed.version != echConfigVersion {
		return "", false
	}

	return string(parsed.publicName), true
}

// ECHConfigToConfigList wraps a single ECHConfig into an ECHConfigList.
func ECHConfigToConfigList(config []byte) ([]byte, error) {
	if len(config) == 0 {
		return nil, errors.New("ECH configuration is empty")
	}
	if len(config) > int(^uint16(0)) {
		return nil, fmt.Errorf("ECH configuration exceeds maximum length of %d bytes", ^uint16(0))
	}

	configList := make([]byte, 2+len(config))
	binary.BigEndian.PutUint16(configList, uint16(len(config)))
	copy(configList[2:], config)

	return configList, nil
}

func parseECHConfigList(configList []byte) ([][]byte, error) {
	if len(configList) < 2 {
		return nil, errors.New("invalid ECH configuration list: expected at least 2 bytes")
	}

	declaredLength := int(binary.BigEndian.Uint16(configList))
	actualLength := len(configList) - 2
	if declaredLength != actualLength {
		return nil, fmt.Errorf("invalid ECH configuration list length: declared %d bytes, got %d", declaredLength, actualLength)
	}
	if actualLength == 0 {
		return nil, errors.New("ECH configuration list is empty")
	}

	var configs [][]byte
	remaining := configList[2:]
	for len(remaining) > 0 {
		if len(remaining) < 4 {
			return nil, errors.New("invalid ECH configuration: truncated header")
		}

		configLength := 4 + int(binary.BigEndian.Uint16(remaining[2:]))
		if configLength > len(remaining) {
			return nil, errors.New("invalid ECH configuration: truncated contents")
		}

		configs = append(configs, remaining[:configLength])
		remaining = remaining[configLength:]
	}

	return configs, nil
}

func parseECHConfig(config []byte) (parsedECHConfig, error) {
	var parsed parsedECHConfig
	input := cryptobyte.String(config)
	var contents cryptobyte.String
	if !input.ReadUint16(&parsed.version) || !input.ReadUint16LengthPrefixed(&contents) || !input.Empty() {
		return parsedECHConfig{}, errors.New("invalid ECH configuration encoding")
	}
	if parsed.version != echConfigVersion {
		return parsed, nil
	}

	var configID uint8
	var publicKey, cipherSuites, publicName, extensions cryptobyte.String
	if !contents.ReadUint8(&configID) ||
		!contents.ReadUint16(&parsed.kemID) ||
		!contents.ReadUint16LengthPrefixed(&publicKey) ||
		!contents.ReadUint16LengthPrefixed(&cipherSuites) ||
		!contents.ReadUint8(&parsed.maximumNameLength) ||
		!contents.ReadUint8LengthPrefixed(&publicName) ||
		!contents.ReadUint16LengthPrefixed(&extensions) ||
		!contents.Empty() {
		return parsedECHConfig{}, errors.New("invalid ECH configuration contents")
	}
	if len(publicKey) == 0 {
		return parsedECHConfig{}, errors.New("ECH configuration public key is empty")
	}
	if len(cipherSuites) == 0 || len(cipherSuites)%4 != 0 {
		return parsedECHConfig{}, errors.New("invalid ECH symmetric cipher suites")
	}
	if len(publicName) == 0 {
		return parsedECHConfig{}, errors.New("ECH configuration public name is empty")
	}
	if !validECHPublicName(publicName) {
		return parsedECHConfig{}, fmt.Errorf("invalid ECH configuration public name %q", publicName)
	}

	seenExtensions := make(map[uint16]struct{})
	for !extensions.Empty() {
		var extensionType uint16
		var extensionData cryptobyte.String
		if !extensions.ReadUint16(&extensionType) || !extensions.ReadUint16LengthPrefixed(&extensionData) {
			return parsedECHConfig{}, errors.New("invalid ECH configuration extensions")
		}
		// Clients ignore configurations with unsupported mandatory extensions (RFC 9849, Section 4.2).
		if extensionType&0x8000 != 0 {
			return parsedECHConfig{}, fmt.Errorf("%w 0x%04x", errUnsupportedMandatoryECHConfigExtension, extensionType)
		}
		if _, duplicate := seenExtensions[extensionType]; duplicate {
			return parsedECHConfig{}, fmt.Errorf("duplicate ECH configuration extension 0x%04x", extensionType)
		}
		seenExtensions[extensionType] = struct{}{}
	}

	parsed.publicKey = publicKey
	parsed.publicName = publicName

	return parsed, nil
}

func validECHPublicName(name []byte) bool {
	// Stricter than RFC 9849's 255-byte bound: Go clients skip names above the 253-byte DNS limit or with a single label.
	if len(name) > maxECHPublicNameLength || !bytes.ContainsRune(name, '.') {
		return false
	}

	labels := bytes.Split(name, []byte{'.'})
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if (c < '0' || c > '9') && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && c != '-' {
				return false
			}
		}
	}

	// Reject IPv4-shaped names: an all-digit final label, or "0x" plus hexadecimal digits (RFC 9849, Section 6.1.7).
	finalLabel := labels[len(labels)-1]
	if len(bytes.Trim(finalLabel, "0123456789")) == 0 {
		return false
	}
	if len(finalLabel) >= 2 && bytes.EqualFold(finalLabel[:2], []byte("0x")) && len(bytes.Trim(finalLabel[2:], "0123456789abcdefABCDEF")) == 0 {
		return false
	}

	return true
}
