package safe

import (
	"fmt"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/log"
	"golang.org/x/net/context"
	"runtime/debug"
	"sync"
)

type routine struct {
	goroutine func(chan bool)
	stop      chan bool
}

type routineCtx func(ctx context.Context)

// Pool is a pool of go routines
type Pool struct {
	routines    []routine
	routinesCtx []routineCtx
	waitGroup   sync.WaitGroup
	lock        sync.Mutex
	baseCtx     context.Context
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewPool creates a Pool
func NewPool(parentCtx context.Context) *Pool {
	baseCtx, _ := context.WithCancel(parentCtx)
	ctx, cancel := context.WithCancel(baseCtx)
	return &Pool{
		baseCtx: baseCtx,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Ctx returns main context
func (p *Pool) Ctx() context.Context {
	return p.baseCtx
}

//AddGoCtx adds a recoverable goroutine with a context without starting it
func (p *Pool) AddGoCtx(goroutine routineCtx) {
	p.lock.Lock()
	p.routinesCtx = append(p.routinesCtx, goroutine)
	p.lock.Unlock()
}

//GoCtx starts a recoverable goroutine with a context
func (p *Pool) GoCtx(goroutine routineCtx) {
	p.lock.Lock()
	p.routinesCtx = append(p.routinesCtx, goroutine)
	p.waitGroup.Add(1)
	Go(func() {
		goroutine(p.ctx)
		p.waitGroup.Done()
	})
	p.lock.Unlock()
}

// Go starts a recoverable goroutine, and can be stopped with stop chan
func (p *Pool) Go(goroutine func(stop chan bool)) {
	p.lock.Lock()
	newRoutine := routine{
		goroutine: goroutine,
		stop:      make(chan bool, 1),
	}
	p.routines = append(p.routines, newRoutine)
	p.waitGroup.Add(1)
	Go(func() {
		goroutine(newRoutine.stop)
		p.waitGroup.Done()
	})
	p.lock.Unlock()
}

// Stop stops all started routines, waiting for their termination
func (p *Pool) Stop() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.cancel()
	for _, routine := range p.routines {
		routine.stop <- true
	}
	p.waitGroup.Wait()
	for _, routine := range p.routines {
		close(routine.stop)
	}
}

// Start starts all stopped routines
func (p *Pool) Start() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.ctx, p.cancel = context.WithCancel(p.baseCtx)
	for _, routine := range p.routines {
		p.waitGroup.Add(1)
		routine.stop = make(chan bool, 1)
		Go(func() {
			routine.goroutine(routine.stop)
			p.waitGroup.Done()
		})
	}

	for _, routine := range p.routinesCtx {
		p.waitGroup.Add(1)
		Go(func() {
			routine(p.ctx)
			p.waitGroup.Done()
		})
	}
}

// Go starts a recoverable goroutine
func Go(goroutine func()) {
	GoWithRecover(goroutine, defaultRecoverGoroutine)
}

// GoWithRecover starts a recoverable goroutine using given customRecover() function
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
	log.Errorf("Error in Go routine: %s", err)
	debug.PrintStack()
}

// OperationWithRecover wrap a backoff operation in a Recover
func OperationWithRecover(operation backoff.Operation) backoff.Operation {
	return func() (err error) {
		defer func() {
			if res := recover(); res != nil {
				defaultRecoverGoroutine(res)
				err = fmt.Errorf("Panic in operation: %s", err)
			}
		}()
		return operation()
	}
}
