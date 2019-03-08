package integration

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/testhelpers"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
	v1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sSuite
type K8sSuite struct{ BaseSuite }

const (
	kubeServer = "https://127.0.0.1:6443"
	namespace  = "default"
)

func (s *K8sSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "k8s")
	s.composeProject.Start(c)
}

func (s *K8sSuite) TearDownSuite(c *check.C) {
	s.composeProject.Stop(c)
	os.Remove("./resources/compose/output/kubeconfig.yaml")
}

func parseK8sYaml(fileR []byte) []runtime.Object {
	acceptedK8sTypes := regexp.MustCompile(`(Deployment|Service|Ingress)`)
	sepYamlfiles := strings.Split(string(fileR), "---")
	retVal := make([]runtime.Object, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(f), nil, nil)

		if err != nil {
			log.WithoutContext().Debugf("Error while decoding YAML object. Err was: %s", err)
			continue
		}

		if !acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
			log.WithoutContext().Debugf("The custom-roles configMap contained K8s object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind)
		} else {
			retVal = append(retVal, obj)
		}
	}
	return retVal
}

func (s *K8sSuite) TestSimpleDefaultConfig(c *check.C) {
	req := testhelpers.MustNewRequest(http.MethodGet, kubeServer, nil)
	err := try.RequestWithTransport(req, time.Second*60, &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, try.StatusCodeIs(http.StatusUnauthorized))
	c.Assert(err, checker.IsNil)

	abs, err := filepath.Abs("./resources/compose/output/kubeconfig.yaml")
	c.Assert(err, checker.IsNil)

	err = try.Do(time.Second*60, try.DoCondition(func() error {
		_, err := os.Stat(abs)
		return err
	}))
	c.Assert(err, checker.IsNil)

	err = os.Setenv("KUBECONFIG", abs)
	c.Assert(err, checker.IsNil)

	cmd, display := s.traefikCmd(withConfigFile("fixtures/k8s_default.toml"))
	defer display(c)

	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	config, err := clientcmd.BuildConfigFromFlags("", abs)
	c.Assert(err, checker.IsNil)

	clientset, err := kubernetes.NewForConfig(config)
	c.Assert(err, checker.IsNil)

	yamlContent, err := ioutil.ReadFile("./fixtures/k8s/test.yml")
	c.Assert(err, checker.IsNil)

	k8sObjects := parseK8sYaml(yamlContent)
	for _, obj := range k8sObjects {
		switch o := obj.(type) {
		case *v1beta12.Deployment:
			_, err := clientset.ExtensionsV1beta1().Deployments(namespace).Create(o)
			c.Assert(err, checker.IsNil)
		case *v1.Service:
			_, err := clientset.CoreV1().Services(namespace).Create(o)
			c.Assert(err, checker.IsNil)
		case *v1beta12.Ingress:
			_, err := clientset.ExtensionsV1beta1().Ingresses(namespace).Create(o)
			c.Assert(err, checker.IsNil)
		default:
			log.WithoutContext().Errorf("Unknown runtime object %+v %T", o, o)
		}

	}

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/kubernetes/routers", 60*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("Host(`whoami.test`)"))
	c.Assert(err, checker.IsNil)
}
