package supervisor

import (
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/timetools"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/proxy"
)

// Supervisor watches changes to the dynamic backends and applies those changes to the server in real time.
// Supervisor handles lifetime of the proxy as well, and does graceful restarts and recoveries in case of failures.
type Supervisor struct {
	// lastId allows to create iterative server instance versions for debugging purposes.
	lastId int

	// wg allows to wait for graceful shutdowns
	wg *sync.WaitGroup

	mtx *sync.RWMutex

	// srv is the current active server
	proxy proxy.Proxy

	// newProxy returns new server instance every time is called.
	newProxy proxy.NewProxyFn

	// timeProvider is used to mock time in tests
	timeProvider timetools.TimeProvider

	// engine is used for reading configuration details
	engine engine.Engine

	// errorC is a channel will be used to notify the calling party of the errors.
	errorC chan error
	// restartC channel is used internally to trigger graceful restarts on errors and configuration changes.
	restartC chan error
	// closeC is a channel to tell everyone to stop working and exit at the earliest convenience.
	closeC chan bool
	// broadcastCloseC is a channel to broadcast the beginning of a close.
	broadcastCloseC chan bool

	options Options

	state supervisorState
}

type Options struct {
	Clock timetools.TimeProvider
	Files []*proxy.FileDescriptor
}

func New(newProxy proxy.NewProxyFn, engine engine.Engine, errorC chan error, options Options) *Supervisor {
	return &Supervisor{
		wg:              &sync.WaitGroup{},
		mtx:             &sync.RWMutex{},
		newProxy:        newProxy,
		engine:          engine,
		options:         setDefaults(options),
		errorC:          errorC,
		restartC:        make(chan error),
		closeC:          make(chan bool),
		broadcastCloseC: make(chan bool, 10),
	}
}

func (s *Supervisor) String() string {
	return fmt.Sprintf("Supervisor(%v)", s.state)
}

func (s *Supervisor) getCurrentProxy() proxy.Proxy {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.proxy
}

func (s *Supervisor) setCurrentProxy(p proxy.Proxy) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.proxy = p
}

func (s *Supervisor) GetFiles() ([]*proxy.FileDescriptor, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.proxy != nil {
		return s.proxy.GetFiles()
	}
	return []*proxy.FileDescriptor{}, nil
}

func (s *Supervisor) setState(state supervisorState) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.state = state
}

func (s *Supervisor) FrontendStats(key engine.FrontendKey) (*engine.RoundTripStats, error) {
	p := s.getCurrentProxy()
	if p != nil {
		return p.FrontendStats(key)
	}
	return nil, fmt.Errorf("no current proxy")
}

func (s *Supervisor) ServerStats(key engine.ServerKey) (*engine.RoundTripStats, error) {
	p := s.getCurrentProxy()
	if p != nil {
		return p.ServerStats(key)
	}
	return nil, fmt.Errorf("no current proxy")
}

func (s *Supervisor) BackendStats(key engine.BackendKey) (*engine.RoundTripStats, error) {
	p := s.getCurrentProxy()
	if p != nil {
		return p.BackendStats(key)
	}
	return nil, fmt.Errorf("no current proxy")
}

// TopFrontends returns locations sorted by criteria (faulty, slow, most used)
// if hostname or backendId is present, will filter out locations for that host or backendId
func (s *Supervisor) TopFrontends(key *engine.BackendKey) ([]engine.Frontend, error) {
	p := s.getCurrentProxy()
	if p != nil {
		return p.TopFrontends(key)
	}
	return nil, fmt.Errorf("no current proxy")
}

// TopServers returns endpoints sorted by criteria (faulty, slow, mos used)
// if backendId is not empty, will filter out endpoints for that backendId
func (s *Supervisor) TopServers(key *engine.BackendKey) ([]engine.Server, error) {
	p := s.getCurrentProxy()
	if p != nil {
		return p.TopServers(key)
	}
	return nil, fmt.Errorf("no current proxy")
}

func (s *Supervisor) init() error {
	proxy, err := s.newProxy(s.lastId)
	s.lastId += 1
	if err != nil {
		return err
	}

	stopNewProxy := true

	defer func() {
		if stopNewProxy {
			proxy.Stop(true)
		}
	}()

	if err := initProxy(s.engine, proxy); err != nil {
		return err
	}

	// This is the first start, pass the files that could have been passed
	// to us by the parent process
	if s.lastId == 1 && len(s.options.Files) != 0 {
		log.Infof("Passing files %v to %v", s.options.Files, proxy)
		if err := proxy.TakeFiles(s.options.Files); err != nil {
			return err
		}
	}

	log.Infof("%v init() initial setup done", proxy)

	oldProxy := s.getCurrentProxy()
	if oldProxy != nil {
		files, err := oldProxy.GetFiles()
		if err != nil {
			return err
		}
		log.Infof("%v taking files from %v to %v", s, oldProxy, proxy)
		if err := proxy.TakeFiles(files); err != nil {
			return err
		}
	}

	if err := proxy.Start(); err != nil {
		return err
	}

	if oldProxy != nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			oldProxy.Stop(true)
		}()
	}

	// Watch and configure this instance of server
	stopNewProxy = false
	s.setCurrentProxy(proxy)
	changesC := make(chan interface{})

	// This goroutine will connect to the backend and emit the changes to the changesC channel.
	// In case of any error it notifies supervisor of the error by sending an error to the channel triggering reload.
	go func() {
		cancelC := make(chan bool)
		if err := s.engine.Subscribe(changesC, cancelC); err != nil {
			log.Infof("%v engine watcher got error: '%v' will restart", proxy, err)
			close(cancelC)
			close(changesC)
			s.restartC <- err
		} else {
			close(cancelC)
			// Graceful shutdown without restart
			log.Infof("%v engine watcher got nil error, gracefully shutdown", proxy)
			s.broadcastCloseC <- true
		}
	}()

	// This goroutine will listen for changes arriving to the changes channel and reconfigure the given server
	go func() {
		for {
			change := <-changesC
			if change == nil {
				log.Infof("Stop watching changes for %s", proxy)
				return
			}
			if err := processChange(proxy, change); err != nil {
				log.Errorf("failed to process change %#v, err: %s", change, err)
			}
		}
	}()
	return nil
}

func (s *Supervisor) stop() {
	srv := s.getCurrentProxy()
	if srv != nil {
		srv.Stop(true)
		log.Infof("%s was stopped by supervisor", srv)
	}
	log.Infof("Wait for any outstanding operations to complete")
	s.wg.Wait()
	log.Infof("All outstanding operations have been completed, signalling stop")
	close(s.closeC)
}

// supervise() listens for error notifications and triggers graceful restart
func (s *Supervisor) supervise() {
	for {
		select {
		case err := <-s.restartC:
			// This means graceful shutdown, do nothing and return
			if err == nil {
				log.Infof("watchErrors - graceful shutdown")
				s.stop()
				return
			}
			for {
				s.options.Clock.Sleep(retryPeriod)
				log.Infof("supervise() restarting %s on error: %s", s.proxy, err)
				// We failed to initialize server, this error can not be recovered, so send an error and exit
				if err := s.init(); err != nil {
					log.Infof("Failed to initialize %s, will retry", err)
				} else {
					break
				}
			}

		case <-s.broadcastCloseC:
			s.Stop(false)
		}
	}
}

func (s *Supervisor) Start() error {
	if s.checkAndSetState(supervisorStateActive) {
		return fmt.Errorf("%v already started", s)
	}
	defer s.setState(supervisorStateActive)
	go s.supervise()
	return s.init()
}

func (s *Supervisor) checkAndSetState(state supervisorState) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.state == state {
		return true
	}

	s.state = state
	return false
}

func (s *Supervisor) Stop(wait bool) {

	// It was already stopped
	if s.checkAndSetState(supervisorStateStopped) {
		return
	}

	close(s.restartC)
	if wait {
		<-s.closeC
		log.Infof("All operations stopped")
	}
}

// initProxy reads the configuration from the engine and configures the server
func initProxy(ng engine.Engine, p proxy.Proxy) error {
	hosts, err := ng.GetHosts()
	if err != nil {
		return err
	}

	for _, h := range hosts {
		if err := p.UpsertHost(h); err != nil {
			return err
		}
	}

	bs, err := ng.GetBackends()
	if err != nil {
		return err
	}

	for _, b := range bs {
		if err := p.UpsertBackend(b); err != nil {
			return err
		}

		bk := engine.BackendKey{Id: b.Id}
		servers, err := ng.GetServers(bk)
		if err != nil {
			return err
		}

		for _, s := range servers {
			if err := p.UpsertServer(bk, s); err != nil {
				return err
			}
		}
	}

	ls, err := ng.GetListeners()
	if err != nil {
		return err
	}

	for _, l := range ls {
		if err := p.UpsertListener(l); err != nil {
			return err
		}
	}

	fs, err := ng.GetFrontends()
	if err != nil {
		return err
	}

	if len(fs) == 0 {
		log.Warningf("No frontends found")
	}

	for _, f := range fs {
		if err := p.UpsertFrontend(f); err != nil {
			return err
		}
		fk := engine.FrontendKey{Id: f.Id}
		ms, err := ng.GetMiddlewares(fk)
		if err != nil {
			return err
		}
		for _, m := range ms {
			if err := p.UpsertMiddleware(fk, m); err != nil {
				return err
			}
		}
	}
	return nil
}

func setDefaults(o Options) Options {
	if o.Clock == nil {
		o.Clock = &timetools.RealTime{}
	}
	return o
}

// processChange takes the backend change notification emitted by the backend and applies it to the server
func processChange(p proxy.Proxy, ch interface{}) error {
	switch change := ch.(type) {
	case *engine.HostUpserted:
		return p.UpsertHost(change.Host)
	case *engine.HostDeleted:
		return p.DeleteHost(change.HostKey)

	case *engine.ListenerUpserted:
		return p.UpsertListener(change.Listener)

	case *engine.ListenerDeleted:
		return p.DeleteListener(change.ListenerKey)

	case *engine.FrontendUpserted:
		return p.UpsertFrontend(change.Frontend)
	case *engine.FrontendDeleted:
		return p.DeleteFrontend(change.FrontendKey)

	case *engine.MiddlewareUpserted:
		return p.UpsertMiddleware(change.FrontendKey, change.Middleware)

	case *engine.MiddlewareDeleted:
		return p.DeleteMiddleware(change.MiddlewareKey)

	case *engine.BackendUpserted:
		return p.UpsertBackend(change.Backend)
	case *engine.BackendDeleted:
		return p.DeleteBackend(change.BackendKey)

	case *engine.ServerUpserted:
		return p.UpsertServer(change.BackendKey, change.Server)
	case *engine.ServerDeleted:
		return p.DeleteServer(change.ServerKey)
	}
	return fmt.Errorf("unsupported change: %#v", ch)
}

const retryPeriod = 5 * time.Second

type supervisorState int

const (
	supervisorStateCreated = iota
	supervisorStateActive
	supervisorStateStopped
)

func (s supervisorState) String() string {
	switch s {
	case supervisorStateCreated:
		return "created"
	case supervisorStateActive:
		return "active"
	case supervisorStateStopped:
		return "stopped"
	default:
		return "unkown"
	}
}
