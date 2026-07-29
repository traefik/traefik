package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/api"
)

// GatewayAPISuite tests the Kubernetes Gateway API provider against the
// supported Gateway API CRD versions.
type GatewayAPISuite struct {
	BaseSuite
}

func TestGatewayAPISuite(t *testing.T) {
	suite.Run(t, new(GatewayAPISuite))
}

func (s *GatewayAPISuite) TestGatewayConfiguration() {
	testCases := []struct {
		crdVersion  string
		crdManifest string
		tcpRoute    string
	}{
		{
			crdVersion:  "v1.5.1",
			crdManifest: "./fixtures/k8s-gateway/00-experimental-v1.5.1.yml",
			tcpRoute:    "./fixtures/k8s-gateway/03-tcproute-v1alpha2.yml",
		},
		{
			crdVersion:  "v1.6.1",
			crdManifest: "./fixtures/k8s-gateway/00-experimental-v1.6.1.yml",
			tcpRoute:    "./fixtures/k8s-gateway/03-tcproute-v1.yml",
		},
	}

	for _, test := range testCases {
		s.Run(test.crdVersion, func() {
			ctx := s.T().Context()

			k3sContainer, err := k3s.Run(ctx, k3sImage,
				k3s.WithManifest(test.crdManifest),
				k3s.WithManifest("./fixtures/k8s-gateway/01-services.yml"),
				k3s.WithManifest("./fixtures/k8s-gateway/02-gateway.yml"),
				k3s.WithManifest(test.tcpRoute),
			)
			require.NoError(s.T(), err)

			s.T().Cleanup(func() {
				// The test context is already cancelled at cleanup time.
				if err := k3sContainer.Terminate(context.Background()); err != nil {
					log.Warn().Err(err).Send()
				}
			})

			kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
			require.NoError(s.T(), err)

			kubeconfigPath := filepath.Join(s.T().TempDir(), "kubeconfig.yaml")
			require.NoError(s.T(), os.WriteFile(kubeconfigPath, kubeConfigYaml, 0o644))
			s.T().Setenv("KUBECONFIG", kubeconfigPath)

			s.traefikCmd(withConfigFile("fixtures/k8s_gateway.toml"))

			// The same dynamic configuration is expected for both CRD versions.
			s.testConfiguration("testdata/rawdata-gateway.json", "8080")
		})
	}
}

func (s *GatewayAPISuite) testConfiguration(path, apiPort string) {
	err := try.GetRequest("http://127.0.0.1:"+apiPort+"/api/entrypoints", 20*time.Second, try.BodyContains(`"name":"web"`))
	require.NoError(s.T(), err)

	expectedJSON := filepath.FromSlash(path)

	if *updateExpected {
		fi, err := os.Create(expectedJSON)
		require.NoError(s.T(), err)
		err = fi.Close()
		require.NoError(s.T(), err)
	}

	var buf bytes.Buffer
	err = try.GetRequest("http://127.0.0.1:"+apiPort+"/api/rawdata", 1*time.Minute, try.StatusCodeIs(http.StatusOK), matchesConfig(expectedJSON, &buf))

	if !*updateExpected {
		require.NoError(s.T(), err)
		return
	}

	if err != nil {
		log.Info().Msgf("In file update mode, got expected error: %v", err)
	}

	var rtRepr api.RunTimeRepresentation
	err = json.Unmarshal(buf.Bytes(), &rtRepr)
	require.NoError(s.T(), err)

	newJSON, err := json.MarshalIndent(rtRepr, "", "\t")
	require.NoError(s.T(), err)

	err = os.WriteFile(expectedJSON, newJSON, 0o644)
	require.NoError(s.T(), err)

	s.T().Fatal("We do not want a passing test in file update mode")
}
