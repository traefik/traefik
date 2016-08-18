package safe

import (
	"golang.org/x/net/context"
	"log"
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
func NewPool(baseCtx context.Context) *Pool {
	ctx, cancel := context.WithCancel(baseCtx)
	return &Pool{
		ctx:     ctx,
		cancel:  cancel,
		baseCtx: baseCtx,
	}
}

// Ctx returns main context
func (p *Pool) Ctx() context.Context {
	return p.ctx
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
	p.cancel()
	for _, routine := range p.routines {
		routine.stop <- true
	}
	p.waitGroup.Wait()
	for _, routine := range p.routines {
		close(routine.stop)
	}
	p.lock.Unlock()
}

// Start starts all stoped routines
func (p *Pool) Start() {
	p.lock.Lock()
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
	p.lock.Unlock()
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
	log.Println(err)
	debug.PrintStack()
}
