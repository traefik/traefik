package proxy

import (
	"crypto/tls"

	"fmt"
	"golang.org/x/crypto/ocsp"
	"net"
	"net/http"

	"github.com/vulcand/vulcand/engine"

	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/manners"
	"github.com/vulcand/route"
)

// srv contains all that is necessary to run the HTTP(s) server. server does not work on its own,
// it heavily depends on MuxServer and acts as its internal data structure.
type srv struct {
	defaultHost string
	mux         *mux
	srv         *manners.GracefulServer
	proxy       http.Handler
	listener    engine.Listener
	options     Options
	state       int
}

func (s *srv) GetFile() (*FileDescriptor, error) {
	if !s.hasListeners() || s.srv == nil {
		return nil, nil
	}
	file, err := s.srv.GetFile()
	if err != nil {
		return nil, err
	}
	return &FileDescriptor{
		File:    file,
		Address: s.listener.Address,
	}, nil
}

func (s *srv) String() string {
	return fmt.Sprintf("%s->srv(%v, %v)", s.mux, s.state, &s.listener)
}

func newSrv(m *mux, l engine.Listener) (*srv, error) {
	defaultHost := ""
	keyPairs := make(map[engine.HostKey]engine.KeyPair)
	for hk, h := range m.hosts {
		if h.Settings.KeyPair != nil {
			keyPairs[hk] = *h.Settings.KeyPair
		}
		if h.Settings.Default {
			defaultHost = hk.Name
		}
	}
	h, err := scopedHandler(l.Scope, m.router)
	if err != nil {
		return nil, err
	}
	return &srv{
		mux:         m,
		proxy:       h,
		listener:    l,
		defaultHost: defaultHost,
		state:       srvStateInit,
	}, nil
}

func (s *srv) isTLS() bool {
	return s.listener.Protocol == engine.HTTPS
}

func (s *srv) updateListener(l engine.Listener) error {
	// We can not listen for different protocols on the same socket
	if s.listener.Protocol != l.Protocol {
		return fmt.Errorf("conflicting protocol %s and %s", s.listener.Protocol, l.Protocol)
	}
	if l.Scope == s.listener.Scope && (&l).SettingsEquals(&s.listener) {
		return nil
	}

	log.Infof("%v update %v", s, &l)
	handler, err := scopedHandler(l.Scope, s.mux.router)
	if err != nil {
		return err
	}
	s.proxy = handler
	s.listener = l

	return s.reload()
}

func (s *srv) isServing() bool {
	return s.state == srvStateActive
}

func (s *srv) hasListeners() bool {
	return s.state == srvStateActive || s.state == srvStateHijacked
}

func (s *srv) takeFile(f *FileDescriptor) error {
	log.Infof("%s takeFile %v", s, f)

	listener, err := f.ToListener()
	if err != nil {
		return err
	}

	if s.isTLS() {
		tcpListener, ok := listener.(*net.TCPListener)
		if !ok {
			return fmt.Errorf(`%s failed to take file descriptor - it is running in TLS mode so I need a TCP listener, 
but the file descriptor that was given corresponded to a listener of type %T. More about file descriptor: %s`, listener, s, f)
		}
		config, err := s.newTLSConfig()
		if err != nil {
			return err
		}
		listener = manners.NewTLSListener(
			manners.TCPKeepAliveListener{tcpListener}, config)
	}

	s.srv = manners.NewWithOptions(
		manners.Options{
			Server:       s.newHTTPServer(),
			Listener:     listener,
			StateHandler: s.mux.connTracker.RegisterStateChange,
		})
	s.state = srvStateHijacked
	return nil
}

func (s *srv) newHTTPServer() *http.Server {
	return &http.Server{
		Handler:        s.proxy,
		ReadTimeout:    s.options.ReadTimeout,
		WriteTimeout:   s.options.WriteTimeout,
		MaxHeaderBytes: s.options.MaxHeaderBytes,
	}
}

func (s *srv) reload() error {
	if !s.isServing() {
		return nil
	}

	var config *tls.Config

	if s.isTLS() {
		cfg, err := s.newTLSConfig()
		if err != nil {
			return err
		}
		config = cfg
	}

	gracefulServer, err := s.srv.HijackListener(s.newHTTPServer(), config)
	if err != nil {
		return err
	}
	go s.serve(gracefulServer)

	s.srv.Close()
	s.srv = gracefulServer
	return nil
}

func (s *srv) shutdown() {
	if s.srv != nil {
		s.srv.Close()
	}
}

func (s *srv) newTLSConfig() (*tls.Config, error) {
	config, err := s.listener.TLSConfig()
	if err != nil {
		return nil, err
	}

	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	pairs := map[string]tls.Certificate{}
	for _, host := range s.mux.hosts {
		c := host.Settings.KeyPair
		if c == nil {
			continue
		}
		keyPair, err := tls.X509KeyPair(c.Cert, c.Key)
		if err != nil {
			return nil, err
		}
		if host.Settings.OCSP.Enabled {
			log.Infof("%v OCSP is enabled for %v, resolvers: %v", s, host, host.Settings.OCSP.Responders)
			r, err := s.mux.stapler.StapleHost(&host)
			if err != nil {
				log.Warningf("%v failed to staple %v, error %v", s, host, err)
			} else if r.Response.Status == ocsp.Good || r.Response.Status == ocsp.Revoked {
				keyPair.OCSPStaple = r.Staple
			} else {
				log.Warningf("%s got undefined status from OCSP responder: %v", s, r.Response.Status)
			}
		}
		pairs[host.Name] = keyPair
	}

	config.Certificates = make([]tls.Certificate, 0, len(pairs))
	if s.defaultHost != "" {
		keyPair, exists := pairs[s.defaultHost]
		if !exists {
			return nil, fmt.Errorf("default host '%s' certificate is not passed", s.defaultHost)
		}
		config.Certificates = append(config.Certificates, keyPair)
	}

	for h, keyPair := range pairs {
		if h != s.defaultHost {
			config.Certificates = append(config.Certificates, keyPair)
		}
	}

	config.BuildNameToCertificate()
	return config, nil
}

func (s *srv) start() error {
	log.Infof("%s start", s)
	switch s.state {
	case srvStateInit:
		listener, err := net.Listen(s.listener.Address.Network, s.listener.Address.Address)
		if err != nil {
			return err
		}

		if s.isTLS() {
			config, err := s.newTLSConfig()
			if err != nil {
				return err
			}
			listener = manners.NewTLSListener(
				manners.TCPKeepAliveListener{listener.(*net.TCPListener)}, config)
		}
		s.srv = manners.NewWithOptions(
			manners.Options{
				Server:       s.newHTTPServer(),
				Listener:     listener,
				StateHandler: s.mux.connTracker.RegisterStateChange,
			})
		s.state = srvStateActive
		go s.serve(s.srv)
		return nil
	case srvStateHijacked:
		s.state = srvStateActive
		go s.serve(s.srv)
		return nil
	}
	return fmt.Errorf("%v Calling start in unsupported state", s)
}

func (s *srv) serve(srv *manners.GracefulServer) {
	log.Infof("%s serve", s)

	s.mux.wg.Add(1)
	defer s.mux.wg.Done()

	srv.ListenAndServe()

	log.Infof("%v stop", s)
}

type srvState int

const (
	srvStateInit     = iota // server has been created
	srvStateActive   = iota // server is active and is serving requests
	srvStateHijacked = iota // server has hijacked listeners from other server
)

func (s srvState) String() string {
	switch s {
	case srvStateInit:
		return "init"
	case srvStateActive:
		return "active"
	case srvStateHijacked:
		return "hijacked"
	}
	return "undefined"
}

func scopedHandler(scope string, proxy http.Handler) (http.Handler, error) {
	if scope == "" {
		return proxy, nil
	}
	mux := route.NewMux()
	mux.SetNotFound(&DefaultNotFound{})
	if err := mux.Handle(scope, proxy); err != nil {
		return nil, err
	}
	return mux, nil
}
