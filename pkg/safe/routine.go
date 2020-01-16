package safe

import (
	"context"
	"crypto/rand"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/cenkalti/backoff/v3"
	"github.com/containous/traefik/v2/pkg/log"
)

type routine struct {
	goroutine func(chan bool)
	stop      chan bool
}

type routineCtx func(ctx context.Context)

// Pool is a pool of go routines
type Pool struct {
	routines    map[string]*routine   // keyed by a randomly generated id
	routinesCtx map[string]routineCtx // keyed by a randomly generated id
	waitGroup   sync.WaitGroup
	lock        sync.Mutex
	baseCtx     context.Context
	baseCancel  context.CancelFunc
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewPool creates a Pool
func NewPool(parentCtx context.Context) *Pool {
	baseCtx, baseCancel := context.WithCancel(parentCtx)
	ctx, cancel := context.WithCancel(baseCtx)
	return &Pool{
		baseCtx:     baseCtx,
		baseCancel:  baseCancel,
		ctx:         ctx,
		cancel:      cancel,
		routinesCtx: make(map[string]routineCtx),
		routines:    make(map[string]*routine),
	}
}

// Ctx returns main context
func (p *Pool) Ctx() context.Context {
	return p.baseCtx
}

func genID() string {
	buf := make([]byte, 20)
	if n, err := rand.Read(buf); err != nil || n != len(buf) {
		log.Errorf("Error getting random bytes for goroutine ID: %v", err)
		return ""
	}
	return fmt.Sprintf("%x", buf)
}

// genUUID generates an ID that is not already in use for a goroutine.
func (p *Pool) genUUID() string {
	var id string
	for {
		id = genID()
		if id == "" {
			continue
		}
		// Ideally, if we had generics, genUUID could take either one of the maps as argument,
		// and we would therefore only check here in one of the two.
		if p.routinesCtx != nil {
			if _, ok := p.routinesCtx[id]; ok {
				continue
			}
		}
		if p.routines != nil {
			if _, ok := p.routines[id]; ok {
				continue
			}
		}
		break
	}
	return id
}

// AddGoCtx adds a recoverable goroutine with a context without starting it
func (p *Pool) AddGoCtx(goroutine routineCtx) {
	p.lock.Lock()
	id := p.genUUID()
	p.routinesCtx[id] = goroutine
	p.lock.Unlock()
}

// GoCtx starts a recoverable goroutine with a context
func (p *Pool) GoCtx(goroutine routineCtx) {
	p.lock.Lock()
	id := p.genUUID()
	p.routinesCtx[id] = goroutine
	p.waitGroup.Add(1)
	Go(func() {
		defer func() {
			p.lock.Lock()
			delete(p.routinesCtx, id)
			p.lock.Unlock()
		}()
		defer p.waitGroup.Done()
		goroutine(p.ctx)
	})
	p.lock.Unlock()
}

// addGo adds a recoverable goroutine, and can be stopped with stop chan
func (p *Pool) addGo(goroutine func(stop chan bool)) {
	p.lock.Lock()
	newRoutine := &routine{
		goroutine: goroutine,
		stop:      make(chan bool, 1),
	}
	id := p.genUUID()
	p.routines[id] = newRoutine
	p.lock.Unlock()
}

// Go starts a recoverable goroutine, and can be stopped with stop chan
func (p *Pool) Go(goroutine func(stop chan bool)) {
	p.lock.Lock()
	newRoutine := &routine{
		goroutine: goroutine,
		stop:      make(chan bool, 1),
	}
	id := p.genUUID()
	p.routines[id] = newRoutine
	p.waitGroup.Add(1)
	Go(func() {
		defer func() {
			p.lock.Lock()
			delete(p.routines, id)
			p.lock.Unlock()
		}()
		defer p.waitGroup.Done()
		goroutine(newRoutine.stop)
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

// Cleanup releases resources used by the pool, and should be called when the pool will no longer be used
func (p *Pool) Cleanup() {
	p.Stop()
	p.lock.Lock()
	defer p.lock.Unlock()
	p.baseCancel()
}

// Start starts all stopped routines
func (p *Pool) Start() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.ctx, p.cancel = context.WithCancel(p.baseCtx)
	for id, routine := range p.routines {
		id := id
		routine := routine
		routine.stop = make(chan bool, 1)
		p.waitGroup.Add(1)

		Go(func() {
			defer func() {
				p.lock.Lock()
				delete(p.routines, id)
				p.lock.Unlock()
			}()
			defer p.waitGroup.Done()
			routine.goroutine(routine.stop)
		})
	}

	for id, routine := range p.routinesCtx {
		p.waitGroup.Add(1)
		id := id
		routine := routine
		Go(func() {
			defer func() {
				p.lock.Lock()
				delete(p.routinesCtx, id)
				p.lock.Unlock()
			}()
			defer p.waitGroup.Done()
			routine(p.ctx)
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
	logger := log.WithoutContext()
	logger.Errorf("Error in Go routine: %s", err)
	logger.Errorf("Stack: %s", debug.Stack())
}

// OperationWithRecover wrap a backoff operation in a Recover
func OperationWithRecover(operation backoff.Operation) backoff.Operation {
	return func() (err error) {
		defer func() {
			if res := recover(); res != nil {
				defaultRecoverGoroutine(res)
				err = fmt.Errorf("panic in operation: %s", err)
			}
		}()
		return operation()
	}
}
