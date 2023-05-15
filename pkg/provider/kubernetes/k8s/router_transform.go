package k8s

import (
	"context"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

type RouterTransform interface {
	Apply(ctx context.Context, rt *dynamic.Router, annotations map[string]string) error
}
