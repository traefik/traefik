package branching

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/containous/alice"
	"github.com/hashicorp/go-bexpr"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
)

const (
	typeName = "Branching"
)

// Branching middleware.
type Branching struct {
	name    string
	next    http.Handler
	branch  http.Handler
	matcher *bexpr.Evaluator
}

type chainBuilder interface {
	BuildChain(ctx context.Context, middlewares []string) *alice.Chain
}

// New creates a branching middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Branching, builder chainBuilder, name string) (http.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName))
	logger.Debug("Creating middleware")

	if config.Chain == nil || len(config.Chain.Middlewares) == 0 {
		return nil, errors.New("empty branch chain")
	}

	eval, err := bexpr.CreateEvaluator(config.Condition)
	if err != nil {
		return nil, fmt.Errorf("failed to create evaluator for expression %q: %w", config.Condition, err)
	}

	chain := builder.BuildChain(ctx, config.Chain.Middlewares)
	branchHandler, err := chain.Then(next)
	if err != nil {
		return nil, fmt.Errorf("failed to create middleware chain %w", err)
	}

	logger.Printf("%s created, matching %q", name, config.Condition)
	return &Branching{
		name:    name,
		next:    next,
		matcher: eval,
		branch:  branchHandler,
	}, nil
}

func (e *Branching) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, err := e.matcher.Evaluate(req)
	if err != nil {
		log.FromContext(middlewares.GetLoggerCtx(context.Background(), e.name, typeName)).Printf("ignoring branch, unable to match request: %v", err)
	}

	if match {
		e.branch.ServeHTTP(rw, req)
		return
	}

	e.next.ServeHTTP(rw, req)
}
