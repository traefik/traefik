package safe

import (
	"log"
	"runtime/debug"
	"sync"
)

type routine struct {
	goroutine func(chan bool)
	stop      chan bool
}

// Pool creates a pool of go routines
type Pool struct {
	routines  []routine
	waitGroup sync.WaitGroup
	lock      sync.Mutex
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
	for _, routine := range p.routines {
		routine.stop <- true
	}
	p.waitGroup.Wait()
	for _, routine := range p.routines {
		close(routine.stop)
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
