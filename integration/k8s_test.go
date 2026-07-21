package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/api"
)

var updateExpected = flag.Bool("update_expected", false, "Update expected files in testdata")

// K8sSuite tests suite.
type K8sSuite struct {
	BaseSuite

	k3sContainer *k3s.K3sContainer
}

func TestK8sSuite(t *testing.T) {
	suite.Run(t, new(K8sSuite))
}

func (s *K8sSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	manifests, err := filepath.Glob("./fixtures/k8s/*.yml")
	require.NoError(s.T(), err)

	opts := make([]testcontainers.ContainerCustomizer, 0, len(manifests))
	for _, m := range manifests {
		opts = append(opts, k3s.WithManifest(m))
	}

	s.k3sContainer, err = k3s.Run(s.T().Context(), k3sImage, opts...)
	require.NoError(s.T(), err)

	kubeConfigYaml, err := s.k3sContainer.GetKubeConfig(s.T().Context())
	require.NoError(s.T(), err)

	kubeconfigPath := filepath.Join(s.T().TempDir(), "kubeconfig.yaml")
	err = os.WriteFile(kubeconfigPath, kubeConfigYaml, 0o644)
	require.NoError(s.T(), err)

	err = os.Setenv("KUBECONFIG", kubeconfigPath)
	require.NoError(s.T(), err)
}

func (s *K8sSuite) TearDownSuite() {
	if s.k3sContainer != nil {
		if s.T().Failed() || *showLog {
			k3sLogs, err := s.k3sContainer.Logs(s.T().Context())
			if err == nil {
				if res, err := io.ReadAll(k3sLogs); err == nil {
					s.T().Log(string(res))
				}
			}
		}

		if err := s.k3sContainer.Terminate(s.T().Context()); err != nil {
			log.Warn().Err(err).Send()
		}
	}

	s.BaseSuite.TearDownSuite()
}

func (s *K8sSuite) TestIngressConfiguration() {
	s.traefikCmd(withConfigFile("fixtures/k8s_default.toml"))

	s.testConfiguration("testdata/rawdata-ingress.json", "8080")
}

func (s *K8sSuite) TestIngressLabelSelector() {
	s.traefikCmd(withConfigFile("fixtures/k8s_ingress_label_selector.toml"))

	s.testConfiguration("testdata/rawdata-ingress-label-selector.json", "8080")
}

func (s *K8sSuite) TestCRDConfiguration() {
	s.traefikCmd(withConfigFile("fixtures/k8s_crd.toml"))

	s.testConfiguration("testdata/rawdata-crd.json", "8000")
}

func (s *K8sSuite) TestCRDLabelSelector() {
	s.traefikCmd(withConfigFile("fixtures/k8s_crd_label_selector.toml"))

	s.testConfiguration("testdata/rawdata-crd-label-selector.json", "8000")
}

func (s *K8sSuite) TestGatewayConfiguration() {
	s.traefikCmd(withConfigFile("fixtures/k8s_gateway.toml"))

	s.testConfiguration("testdata/rawdata-gateway.json", "8080")
}

func (s *K8sSuite) TestIngressClass() {
	s.traefikCmd(withConfigFile("fixtures/k8s_ingressclass.toml"))

	s.testConfiguration("testdata/rawdata-ingressclass.json", "8080")
}

func (s *K8sSuite) TestDisableIngressClassLookup() {
	s.traefikCmd(withConfigFile("fixtures/k8s_ingressclass_disabled.toml"))

	s.testConfiguration("testdata/rawdata-ingressclass-disabled.json", "8080")
}

func (s *K8sSuite) testConfiguration(path, apiPort string) {
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

func matchesConfig(wantConfig string, buf *bytes.Buffer) try.ResponseCondition {
	return func(res *http.Response) error {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if err := res.Body.Close(); err != nil {
			return err
		}

		var obtained api.RunTimeRepresentation
		err = json.Unmarshal(body, &obtained)
		if err != nil {
			return err
		}

		if buf != nil {
			buf.Reset()
			if _, err := io.Copy(buf, bytes.NewReader(body)); err != nil {
				return err
			}
		}

		got, err := json.MarshalIndent(obtained, "", "\t")
		if err != nil {
			return err
		}

		expected, err := os.ReadFile(wantConfig)
		if err != nil {
			return err
		}

		// The pods IPs are dynamic, so we cannot predict them,
		// which is why we have to ignore them in the comparison.
		rxURL := regexp.MustCompile(`"(url|address)":\s+(".*")`)
		sanitizedExpected := rxURL.ReplaceAll(bytes.TrimSpace(expected), []byte(`"$1": "XXXX"`))
		sanitizedGot := rxURL.ReplaceAll(got, []byte(`"$1": "XXXX"`))

		rxServerStatus := regexp.MustCompile(`"(http://)?.*?":\s+(".*")`)
		sanitizedExpected = rxServerStatus.ReplaceAll(sanitizedExpected, []byte(`"XXXX": $1`))
		sanitizedGot = rxServerStatus.ReplaceAll(sanitizedGot, []byte(`"XXXX": $1`))

		if bytes.Equal(sanitizedExpected, sanitizedGot) {
			return nil
		}

		diff := difflib.UnifiedDiff{
			FromFile: "Expected",
			A:        difflib.SplitLines(string(sanitizedExpected)),
			ToFile:   "Got",
			B:        difflib.SplitLines(string(sanitizedGot)),
			Context:  3,
		}

		text, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			return err
		}
		return errors.New(text)
	}
}
