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
	maxECHPublicNameLength = 255
)

type parsedECHConfig struct {
	version           uint16
	kemID             uint16
	publicKey         []byte
	maximumNameLength uint8
	publicName        []byte
}

func UnmarshalECHKey(data []byte) (*tls.EncryptedClientHelloKey, error) {
	keys, err := UnmarshalECHKeys(data)
	if err != nil {
		return nil, err
	}
	if len(keys) != 1 {
		return nil, fmt.Errorf("expected one ECH configuration matching the private key, got %d", len(keys))
	}

	return &keys[0], nil
}

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
	var privateKey, configList []byte
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}

		switch block.Type {
		case "PRIVATE KEY":
			if privateKey != nil {
				return nil, nil, errors.New("multiple ECH private keys in PEM file")
			}
			privateKey = block.Bytes
		case "ECHCONFIG":
			configList = block.Bytes
			data = nil
			continue
		default:
			return nil, nil, fmt.Errorf("unknown PEM block %s", block.Type)
		}

		data = rest
	}

	if len(privateKey) == 0 || len(configList) == 0 {
		return nil, nil, errors.New("missing ECH configuration or private key in PEM file")
	}

	return privateKey, configList, nil
}

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

func NewECHKey(publicName string) (*tls.EncryptedClientHelloKey, error) {
	if len(publicName) == 0 {
		return nil, errors.New("public name is empty")
	}
	if len(publicName) > maxECHPublicNameLength {
		return nil, fmt.Errorf("public name exceeds maximum length of %d bytes", maxECHPublicNameLength)
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

	for !extensions.Empty() {
		var extensionType uint16
		var extensionData cryptobyte.String
		if !extensions.ReadUint16(&extensionType) || !extensions.ReadUint16LengthPrefixed(&extensionData) {
			return parsedECHConfig{}, errors.New("invalid ECH configuration extensions")
		}
	}

	parsed.publicKey = publicKey
	parsed.publicName = publicName

	return parsed, nil
}

func validECHPublicName(name []byte) bool {
	labels := bytes.Split(name, []byte{'.'})
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || !isASCIILetterOrDigit(label[0]) || !isASCIILetterOrDigit(label[len(label)-1]) {
			return false
		}
		if len(label) > 1 {
			for _, character := range label[1 : len(label)-1] {
				if !isASCIILetterOrDigit(character) && character != '-' {
					return false
				}
			}
		}
	}

	finalLabel := labels[len(labels)-1]
	if allASCII(finalLabel, isASCIIDigit) {
		return false
	}
	if len(finalLabel) >= 2 && bytes.EqualFold(finalLabel[:2], []byte("0x")) && allASCII(finalLabel[2:], isASCIIHexDigit) {
		return false
	}

	return true
}

func allASCII(value []byte, predicate func(byte) bool) bool {
	for _, character := range value {
		if !predicate(character) {
			return false
		}
	}
	return true
}

func isASCIILetterOrDigit(character byte) bool {
	return character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || isASCIIDigit(character)
}

func isASCIIDigit(character byte) bool {
	return character >= '0' && character <= '9'
}

func isASCIIHexDigit(character byte) bool {
	return isASCIIDigit(character) || character >= 'a' && character <= 'f' || character >= 'A' && character <= 'F'
}
