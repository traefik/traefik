package conformance

import (
	"knative.dev/networking/test/conformance/ingress"
	"testing"
)

func TestYourIngressConformance(t *testing.T) {
	ingress.RunConformance(t)
}
