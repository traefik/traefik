package safe

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
)

type routineCtx func(ctx context.Context)

// Pool is a pool of go routines.
type Pool struct {
	waitGroup sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewPool creates a Pool.
func NewPool(parentCtx context.Context) *Pool {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Pool{
		ctx:    ctx,
		cancel: cancel,
	}
}

// GoCtx starts a recoverable goroutine with a context.
func (p *Pool) GoCtx(goroutine routineCtx) {
	p.waitGroup.Add(1)
	Go(func() {
		defer p.waitGroup.Done()
		goroutine(p.ctx)
	})
}

// Stop stops all started routines, waiting for their termination.
func (p *Pool) Stop() {
	p.cancel()
	p.waitGroup.Wait()
}

// Go starts a recoverable goroutine.
func Go(goroutine func()) {
	GoWithRecover(goroutine, defaultRecoverGoroutine)
}

// GoWithRecover starts a recoverable goroutine using given customRecover() function.
func GoWithRecover(goroutine func(), customRecover func(err interface{})) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				customRecover(err)
			}
		}()
		goroutine()
	}()
}

func defaultRecoverGoroutine(err interface{}) {
	log.Error().Interface("error", err).Msg("Error in Go routine")
	log.Error().Msgf("Stack: %s", debug.Stack())
}

// OperationWithRecover wrap a backoff operation in a Recover.
func OperationWithRecover(operation backoff.Operation) backoff.Operation {
	return func() (err error) {
		defer func() {
			if res := recover(); res != nil {
				defaultRecoverGoroutine(res)
				err = fmt.Errorf("panic in operation: %w", err)
			}
		}()
		return operation()
	}
}
