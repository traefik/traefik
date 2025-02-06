package conformance

import (
	"flag"
	"testing"

	"knative.dev/networking/test/conformance/ingress"
)

func TestYourIngressConformance(t *testing.T) {
	err := flag.CommandLine.Set("ingressendpoint", "localhost")
	if err != nil {
		return
	}
	ingress.RunConformance(t)
}
