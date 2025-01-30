package conformance

import (
	"testing"

	"knative.dev/networking/test/conformance/ingress"
)

func TestYourIngressConformance(t *testing.T) {
	ingress.RunConformance(t)
}
