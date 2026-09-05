package tls

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
)

const rfc9934ECHKey = `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VuBCIEICjd4yGRdsoP9gU7YT7My8DHx1Tjme8GYDXrOMCi8v1V
-----END PRIVATE KEY-----
-----BEGIN ECHCONFIG-----
AD7+DQA65wAgACA8wVN2BtscOl3vQheUzHeIkVmKIiydUhDCliA4iyQRCwAEAAEA
AQALZXhhbXBsZS5jb20AAA==
-----END ECHCONFIG-----
`

func TestNewECHKey(t *testing.T) {
	testCases := []struct {
		desc        string
		publicName  string
		expectError bool
	}{
		{
			desc:       "valid public name",
			publicName: "server.local",
		},
		{
			desc:       "public name longer than maximum name length",
			publicName: "a-long-public-name-for-ech.example.com",
		},
		{
			desc:       "public name at DNS length limit",
			publicName: strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61),
		},
		{
			desc:        "empty public name",
			expectError: true,
		},
		{
			desc:        "public name exceeds DNS length limit",
			publicName:  strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 62),
			expectError: true,
		},
		{
			desc:        "single-label public name",
			publicName:  "localhost",
			expectError: true,
		},
		{
			desc:        "public name contains empty label",
			publicName:  "server..example.com",
			expectError: true,
		},
		{
			desc:        "public name contains invalid label",
			publicName:  "server_name.example.com",
			expectError: true,
		},
		{
			desc:        "public name is IPv4-like",
			publicName:  "127.0.0.1",
			expectError: true,
		},
		{
			desc:        "public name ends with hexadecimal IPv4-like label",
			publicName:  "server.0xdeadbeef",
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			key, err := NewECHKey(test.publicName)
			if test.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, key.Config)
			assert.Len(t, key.PrivateKey, 32)
			assert.True(t, key.SendAsRetry)

			parsed, err := parseECHConfig(key.Config)
			require.NoError(t, err)
			assert.Zero(t, parsed.maximumNameLength)
			assert.Equal(t, test.publicName, string(parsed.publicName))
		})
	}
}

func TestMarshalUnmarshalECHKey(t *testing.T) {
	key, err := NewECHKey("server.local")
	require.NoError(t, err)

	data, err := MarshalECHKey(key)
	require.NoError(t, err)

	privateKeyBlock, rest := pem.Decode(data)
	require.NotNil(t, privateKeyBlock)
	assert.Equal(t, "PRIVATE KEY", privateKeyBlock.Type)
	parsedPrivateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	require.NoError(t, err)
	assert.IsType(t, &ecdh.PrivateKey{}, parsedPrivateKey)

	configBlock, _ := pem.Decode(rest)
	require.NotNil(t, configBlock)
	assert.Equal(t, "ECHCONFIG", configBlock.Type)
	assert.Equal(t, len(configBlock.Bytes)-2, int(binary.BigEndian.Uint16(configBlock.Bytes)))

	decoded, err := UnmarshalECHKeys(data)
	require.NoError(t, err)
	require.Len(t, decoded, 1)
	assert.Equal(t, key.Config, decoded[0].Config)
	assert.Equal(t, key.PrivateKey, decoded[0].PrivateKey)
	assert.True(t, decoded[0].SendAsRetry)
}

func TestUnmarshalECHKeyRFC9934(t *testing.T) {
	keys, err := UnmarshalECHKeys([]byte(rfc9934ECHKey))
	require.NoError(t, err)
	require.Len(t, keys, 1)

	assert.Len(t, keys[0].PrivateKey, 32)
	assert.True(t, keys[0].SendAsRetry)

	parsed, err := parseECHConfig(keys[0].Config)
	require.NoError(t, err)
	assert.Equal(t, uint16(echConfigVersion), parsed.version)
	assert.Equal(t, uint16(0x0020), parsed.kemID)
	assert.Equal(t, "example.com", string(parsed.publicName))
}

func TestMarshalECHKeyErrors(t *testing.T) {
	key, err := NewECHKey("server.local")
	require.NoError(t, err)

	otherKey, err := NewECHKey("server.local")
	require.NoError(t, err)

	testCases := []struct {
		desc string
		key  *tls.EncryptedClientHelloKey
	}{
		{
			desc: "nil key",
		},
		{
			desc: "missing config",
			key:  &tls.EncryptedClientHelloKey{PrivateKey: key.PrivateKey},
		},
		{
			desc: "missing private key",
			key:  &tls.EncryptedClientHelloKey{Config: key.Config},
		},
		{
			desc: "invalid private key",
			key:  &tls.EncryptedClientHelloKey{Config: key.Config, PrivateKey: []byte("invalid")},
		},
		{
			desc: "invalid config",
			key:  &tls.EncryptedClientHelloKey{Config: []byte("invalid"), PrivateKey: key.PrivateKey},
		},
		{
			desc: "mismatched key",
			key:  &tls.EncryptedClientHelloKey{Config: key.Config, PrivateKey: otherKey.PrivateKey},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := MarshalECHKey(test.key)
			require.Error(t, err)
		})
	}
}

func TestUnmarshalECHKeyErrors(t *testing.T) {
	key, err := NewECHKey("server.local")
	require.NoError(t, err)

	data, err := MarshalECHKey(key)
	require.NoError(t, err)
	privateKeyDER, configList := decodeECHBlocks(t, data)

	otherKey, err := NewECHKey("server.local")
	require.NoError(t, err)
	otherConfigList, err := ECHConfigToConfigList(otherKey.Config)
	require.NoError(t, err)
	invalidNameConfig := append([]byte(nil), key.Config...)
	copy(invalidNameConfig[len(invalidNameConfig)-14:len(invalidNameConfig)-2], "server_local")
	invalidNameConfigList, err := ECHConfigToConfigList(invalidNameConfig)
	require.NoError(t, err)

	singleLabelConfigList, err := ECHConfigToConfigList(configWithPublicName(t, key.Config, "localhost"))
	require.NoError(t, err)

	longName := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 62)
	longNameConfigList, err := ECHConfigToConfigList(configWithPublicName(t, key.Config, longName))
	require.NoError(t, err)

	_, ed25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	ed25519DER, err := x509.MarshalPKCS8PrivateKey(ed25519Key)
	require.NoError(t, err)

	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ecdsaDER, err := x509.MarshalPKCS8PrivateKey(ecdsaKey)
	require.NoError(t, err)

	testCases := []struct {
		desc        string
		data        []byte
		errContains string
	}{
		{
			desc: "empty data",
		},
		{
			desc:        "not PEM content",
			data:        []byte("/etc/traefik/missing-ech.pem"),
			errContains: "no PEM block",
		},
		{
			desc:        "ECHCONFIG block before private key",
			data:        append(encodeECHPEM(nil, configList), encodeECHPEM(privateKeyDER, nil)...),
			errContains: "must precede",
		},
		{
			desc:        "duplicate private keys",
			data:        append(encodeECHPEM(privateKeyDER, nil), encodeECHPEM(privateKeyDER, configList)...),
			errContains: "multiple ECH private keys",
		},
		{
			desc:        "Ed25519 private key",
			data:        encodeECHPEM(ed25519DER, configList),
			errContains: "unsupported ECH private key type",
		},
		{
			desc:        "ECDSA private key",
			data:        encodeECHPEM(ecdsaDER, configList),
			errContains: "unsupported ECH private key type",
		},
		{
			desc: "unknown PEM block",
			data: pem.EncodeToMemory(&pem.Block{Type: "UNKNOWN", Bytes: []byte("data")}),
		},
		{
			desc: "missing private key",
			data: encodeECHPEM(nil, configList),
		},
		{
			desc: "missing config",
			data: encodeECHPEM(privateKeyDER, nil),
		},
		{
			desc: "raw private key instead of PKCS8",
			data: encodeECHPEM(make([]byte, 32), configList),
		},
		{
			desc: "configuration list shorter than length prefix",
			data: encodeECHPEM(privateKeyDER, []byte{0}),
		},
		{
			desc: "configuration list length mismatch",
			data: encodeECHPEM(privateKeyDER, append([]byte{0, 1}, key.Config...)),
		},
		{
			desc: "empty configuration list",
			data: encodeECHPEM(privateKeyDER, []byte{0, 0}),
		},
		{
			desc: "truncated configuration",
			data: encodeECHPEM(privateKeyDER, []byte{0, 4, 0xfe, 0x0d, 0, 1}),
		},
		{
			desc: "configuration does not match key",
			data: encodeECHPEM(privateKeyDER, otherConfigList),
		},
		{
			desc: "invalid public name",
			data: encodeECHPEM(privateKeyDER, invalidNameConfigList),
		},
		{
			desc: "single-label public name",
			data: encodeECHPEM(privateKeyDER, singleLabelConfigList),
		},
		{
			desc: "public name exceeds DNS length limit",
			data: encodeECHPEM(privateKeyDER, longNameConfigList),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := UnmarshalECHKeys(test.data)
			require.Error(t, err)
			if test.errContains != "" {
				assert.ErrorContains(t, err, test.errContains)
			}
		})
	}
}

func TestUnmarshalECHKeys(t *testing.T) {
	key, err := NewECHKey("server.local")
	require.NoError(t, err)

	data, err := MarshalECHKey(key)
	require.NoError(t, err)
	privateKeyDER, _ := decodeECHBlocks(t, data)

	secondConfig := append([]byte(nil), key.Config...)
	secondConfig[4]++
	unknownVersionConfig := append([]byte(nil), key.Config...)
	unknownVersionConfig[1]++
	unsupportedMandatoryConfig := configWithExtensions(t, key.Config, []byte{0x80, 0x01, 0x00, 0x00})
	configList := encodeECHConfigList(t, unknownVersionConfig, unsupportedMandatoryConfig, key.Config, secondConfig)

	keys, err := UnmarshalECHKeys(encodeECHPEM(privateKeyDER, configList))
	require.NoError(t, err)
	require.Len(t, keys, 2)
	assert.Equal(t, key.Config, keys[0].Config)
	assert.Equal(t, secondConfig, keys[1].Config)
	assert.Equal(t, keys[0].PrivateKey, keys[1].PrivateKey)
	assert.True(t, keys[0].SendAsRetry)
	assert.True(t, keys[1].SendAsRetry)
}

func TestUnmarshalECHKeysIgnoresTrailingContent(t *testing.T) {
	key, err := NewECHKey("server.local")
	require.NoError(t, err)

	data, err := MarshalECHKey(key)
	require.NoError(t, err)

	// Content after the ECHConfigList block is ignored (RFC 9934, Section 3).
	data = append(data, pem.EncodeToMemory(&pem.Block{Type: "GARBAGE", Bytes: []byte("garbage")})...)

	keys, err := UnmarshalECHKeys(data)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
}

func TestECHPublicName(t *testing.T) {
	key, err := NewECHKey("server.local")
	require.NoError(t, err)

	name, ok := ECHPublicName(*key)
	require.True(t, ok)
	assert.Equal(t, "server.local", name)

	unknownVersionConfig := append([]byte(nil), key.Config...)
	unknownVersionConfig[1]++
	_, ok = ECHPublicName(tls.EncryptedClientHelloKey{Config: unknownVersionConfig})
	assert.False(t, ok)

	_, ok = ECHPublicName(tls.EncryptedClientHelloKey{Config: []byte("invalid")})
	assert.False(t, ok)
}

func TestParseECHConfigExtensions(t *testing.T) {
	key, err := NewECHKey("server.local")
	require.NoError(t, err)

	testCases := []struct {
		desc        string
		extensions  []byte
		errContains string
	}{
		{
			desc:       "unknown non-mandatory extension",
			extensions: []byte{0x00, 0x01, 0x00, 0x00},
		},
		{
			desc:        "unknown mandatory extension",
			extensions:  []byte{0x80, 0x01, 0x00, 0x00},
			errContains: "mandatory",
		},
		{
			desc:        "duplicate extension",
			extensions:  []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00},
			errContains: "duplicate",
		},
		{
			desc:        "truncated extension",
			extensions:  []byte{0x00, 0x01},
			errContains: "invalid ECH configuration extensions",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := parseECHConfig(configWithExtensions(t, key.Config, test.extensions))
			if test.errContains != "" {
				require.ErrorContains(t, err, test.errContains)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidECHPublicName(t *testing.T) {
	testCases := []struct {
		name  string
		valid bool
	}{
		{name: "example.com", valid: true},
		{name: "EXAMPLE.COM", valid: true},
		{name: "a-b.example.com", valid: true},
		{name: "1.2.3.4.example.com", valid: true},
		{name: strings.Repeat("a", 63) + ".example.com", valid: true},
		{name: strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61), valid: true},
		{name: "server.0xzz", valid: true},
		{name: "server.1e5", valid: true},
		{name: "localhost", valid: false},
		{name: strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 62), valid: false},
		{name: "example.com.", valid: false},
		{name: ".example.com", valid: false},
		{name: strings.Repeat("a", 64) + ".example.com", valid: false},
		{name: "-a.example.com", valid: false},
		{name: "a-.example.com", valid: false},
		{name: "a_b.example.com", valid: false},
		{name: "127.0.0.1", valid: false},
		{name: "example.123", valid: false},
		{name: "server.0x", valid: false},
		{name: "server.0xff", valid: false},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.valid, validECHPublicName([]byte(test.name)))
		})
	}
}

func TestECHConfigToConfigList(t *testing.T) {
	config := []byte{0x01, 0x02, 0x03, 0x04}
	configList, err := ECHConfigToConfigList(config)
	require.NoError(t, err)

	assert.Equal(t, []byte{0, 4, 1, 2, 3, 4}, configList)

	_, err = ECHConfigToConfigList(nil)
	require.Error(t, err)

	maxSizeList, err := ECHConfigToConfigList(make([]byte, 1<<16-1))
	require.NoError(t, err)
	assert.Len(t, maxSizeList, 1<<16+1)

	_, err = ECHConfigToConfigList(make([]byte, 1<<16))
	require.Error(t, err)
}

func TestBuildTLSConfigWithECH(t *testing.T) {
	retiredKey, err := NewECHKey("server.local")
	require.NoError(t, err)
	retiredKeyPEM, err := MarshalECHKey(retiredKey)
	require.NoError(t, err)

	config, err := buildTLSConfig(Options{
		MinVersion:         "VersionTLS13",
		ECHKeys:            []types.FileOrContent{rfc9934ECHKey},
		ECHDecryptOnlyKeys: []types.FileOrContent{types.FileOrContent(retiredKeyPEM)},
	})
	require.NoError(t, err)

	assert.Equal(t, uint16(tls.VersionTLS13), config.MinVersion)
	require.Len(t, config.EncryptedClientHelloKeys, 2)
	assert.True(t, config.EncryptedClientHelloKeys[0].SendAsRetry)
	assert.False(t, config.EncryptedClientHelloKeys[1].SendAsRetry)
}

func TestBuildTLSConfigWithECHErrors(t *testing.T) {
	testCases := []struct {
		desc        string
		options     Options
		errContains string
	}{
		{
			desc: "minVersion below TLS 1.3",
			options: Options{
				MinVersion: "VersionTLS12",
				ECHKeys:    []types.FileOrContent{rfc9934ECHKey},
			},
			errContains: "minVersion",
		},
		{
			desc: "maxVersion below TLS 1.3",
			options: Options{
				MaxVersion: "VersionTLS12",
				ECHKeys:    []types.FileOrContent{rfc9934ECHKey},
			},
			errContains: "maxVersion",
		},
		{
			desc: "invalid key content",
			options: Options{
				ECHKeys: []types.FileOrContent{"not a PEM"},
			},
			errContains: "unmarshalling ECH keys failed",
		},
		{
			desc: "invalid decrypt-only key content",
			options: Options{
				ECHKeys:            []types.FileOrContent{rfc9934ECHKey},
				ECHDecryptOnlyKeys: []types.FileOrContent{"not a PEM"},
			},
			errContains: "unmarshalling ECH keys failed",
		},
		{
			desc: "decrypt-only keys without advertised keys",
			options: Options{
				ECHDecryptOnlyKeys: []types.FileOrContent{rfc9934ECHKey},
			},
			errContains: "at least one echKeys entry",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := buildTLSConfig(test.options)
			require.ErrorContains(t, err, test.errContains)
		})
	}
}

func TestECHIncompatibleOptions(t *testing.T) {
	echOptions, cappedOptions := echIncompatibleOptions(map[string]Options{
		"ech":       {ECHKeys: []types.FileOrContent{rfc9934ECHKey}},
		"capped":    {MaxVersion: "VersionTLS12"},
		"modern":    {MaxVersion: "VersionTLS13"},
		"echCapped": {MaxVersion: "VersionTLS12", ECHKeys: []types.FileOrContent{rfc9934ECHKey}},
	})

	assert.Equal(t, []string{"ech", "echCapped"}, echOptions)
	assert.Equal(t, []string{"capped"}, cappedOptions)
}

func TestRequestWithECH(t *testing.T) {
	const publicName = "server.local"

	key, err := NewECHKey(publicName)
	require.NoError(t, err)

	certificate, err := generateTestCert(publicName)
	require.NoError(t, err)

	configList, err := ECHConfigToConfigList(key.Config)
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		assert.True(t, request.TLS.ECHAccepted)
		_, _ = fmt.Fprint(w, "ECH accepted")
	}))
	server.TLS = &tls.Config{
		Certificates:             []tls.Certificate{certificate},
		MinVersion:               tls.VersionTLS13,
		EncryptedClientHelloKeys: []tls.EncryptedClientHelloKey{*key},
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:                     publicName,
				MinVersion:                     tls.VersionTLS13,
				EncryptedClientHelloConfigList: configList,
				InsecureSkipVerify:             true,
			},
		},
	}
	response, err := client.Get(server.URL)
	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func decodeECHBlocks(t *testing.T, data []byte) ([]byte, []byte) {
	t.Helper()

	privateKeyBlock, rest := pem.Decode(data)
	require.NotNil(t, privateKeyBlock)
	configBlock, _ := pem.Decode(rest)
	require.NotNil(t, configBlock)

	return privateKeyBlock.Bytes, configBlock.Bytes
}

func encodeECHPEM(privateKey, configList []byte) []byte {
	var data []byte
	if privateKey != nil {
		data = append(data, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})...)
	}
	if configList != nil {
		data = append(data, pem.EncodeToMemory(&pem.Block{Type: "ECHCONFIG", Bytes: configList})...)
	}

	return data
}

// configWithPublicName replaces the public name of an ECHConfig built by NewECHKey.
func configWithPublicName(t *testing.T, config []byte, name string) []byte {
	t.Helper()

	// The public name length byte sits at a fixed offset for X25519 configurations.
	const nameOffset = 48
	require.Greater(t, len(config), nameOffset)
	require.LessOrEqual(t, len(name), 255)

	rebuilt := append([]byte(nil), config[:nameOffset]...)
	rebuilt = append(rebuilt, byte(len(name)))
	rebuilt = append(rebuilt, name...)
	rebuilt = append(rebuilt, 0, 0)
	binary.BigEndian.PutUint16(rebuilt[2:], uint16(len(rebuilt)-4))

	return rebuilt
}

// configWithExtensions replaces the empty extensions block of an ECHConfig built by NewECHKey.
func configWithExtensions(t *testing.T, config, extensions []byte) []byte {
	t.Helper()

	require.GreaterOrEqual(t, len(config), 6)
	rebuilt := append([]byte(nil), config[:len(config)-2]...)
	rebuilt = binary.BigEndian.AppendUint16(rebuilt, uint16(len(extensions)))
	rebuilt = append(rebuilt, extensions...)
	binary.BigEndian.PutUint16(rebuilt[2:], uint16(len(rebuilt)-4))

	return rebuilt
}

func encodeECHConfigList(t *testing.T, configs ...[]byte) []byte {
	t.Helper()

	var contents []byte
	for _, config := range configs {
		contents = append(contents, config...)
	}
	require.LessOrEqual(t, len(contents), int(^uint16(0)))

	configList := make([]byte, 2+len(contents))
	binary.BigEndian.PutUint16(configList, uint16(len(contents)))
	copy(configList[2:], contents)

	return configList
}

func generateTestCert(commonName string) (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generating RSA key: %w", err)
	}

	notBefore := time.Now()
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		DNSNames:              []string{commonName, "localhost"},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(24 * time.Hour),
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("creating certificate: %w", err)
	}

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER})
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return tls.X509KeyPair(certificatePEM, privateKeyPEM)
}
