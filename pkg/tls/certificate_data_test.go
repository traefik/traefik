package tls

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendCertificate(t *testing.T) {
	storesCertificates := make(map[string]map[string]*CertificateData)
	data := CertificateData{
		config: &Certificate{
			CertFile: "../../integration/fixtures/https/snitest.com.cert",
			KeyFile:  "../../integration/fixtures/https/snitest.com.key",
		},
	}

	require.NoError(t, data.AppendCertificate(storesCertificates, DefaultTLSStoreName))
	assert.Equal(t, []string{"snitest.com"}, storesCertificates[DefaultTLSStoreName]["snitest.com"].config.SANs)
	assert.False(t, storesCertificates[DefaultTLSStoreName]["snitest.com"].ocsp.MustStaple)

	data = CertificateData{
		config: &Certificate{
			CertFile: "../../integration/fixtures/https/uppercase_wildcard.www.snitest.com.cert",
			KeyFile:  "../../integration/fixtures/https/uppercase_wildcard.www.snitest.com.key",
		},
	}

	require.NoError(t, data.AppendCertificate(storesCertificates, DefaultTLSStoreName))
	assert.Equal(t, []string{"*.www.snitest.com", "foo.www.snitest.com"}, storesCertificates[DefaultTLSStoreName]["*.www.snitest.com,foo.www.snitest.com"].config.SANs)
	assert.False(t, storesCertificates[DefaultTLSStoreName]["*.www.snitest.com,foo.www.snitest.com"].ocsp.MustStaple)
}
