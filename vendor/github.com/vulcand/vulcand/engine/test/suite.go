package test

import (
	"testing"
	"time"

	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/plugin/connlimit"

	. "gopkg.in/check.v1"
)

func TestEtcd(t *testing.T) { TestingT(t) }

type EngineSuite struct {
	Engine   engine.Engine
	ChangesC chan interface{}
}

func (s *EngineSuite) collectChanges(c *C, expected int) []interface{} {
	changes := make([]interface{}, expected)
	for i, _ := range changes {
		select {
		case changes[i] = <-s.ChangesC:
			// successfully collected changes
		case <-time.After(2 * time.Second):
			c.Fatalf("Timeout occured")
		}
	}
	return changes
}

func (s *EngineSuite) expectChanges(c *C, expected ...interface{}) {
	changes := s.collectChanges(c, len(expected))
	for i, ch := range changes {
		c.Assert(ch, DeepEquals, expected[i])
	}
}

func (s *EngineSuite) makeConnLimit(id, variable string, conns int64) engine.Middleware {
	cl, err := connlimit.NewConnLimit(conns, variable)
	if err != nil {
		panic(err)
	}
	return engine.Middleware{
		Id:         id,
		Type:       "connlimit",
		Priority:   1,
		Middleware: cl,
	}
}

func (s *EngineSuite) EmptyParams(c *C) {
	// Empty host operations
	c.Assert(s.Engine.UpsertHost(engine.Host{}), FitsTypeOf, &engine.InvalidFormatError{})
	c.Assert(s.Engine.DeleteHost(engine.HostKey{}), FitsTypeOf, &engine.InvalidFormatError{})

	// Empty listener operations
	c.Assert(s.Engine.UpsertListener(engine.Listener{}), FitsTypeOf, &engine.InvalidFormatError{})
	c.Assert(s.Engine.DeleteListener(engine.ListenerKey{}), FitsTypeOf, &engine.InvalidFormatError{})

	// Empty backend operations
	c.Assert(s.Engine.UpsertBackend(engine.Backend{}), FitsTypeOf, &engine.InvalidFormatError{})
	c.Assert(s.Engine.DeleteBackend(engine.BackendKey{}), FitsTypeOf, &engine.InvalidFormatError{})

	// Empty server operations
	c.Assert(s.Engine.UpsertServer(engine.BackendKey{}, engine.Server{}, 0), FitsTypeOf, &engine.InvalidFormatError{})
	c.Assert(s.Engine.DeleteServer(engine.ServerKey{}), FitsTypeOf, &engine.InvalidFormatError{})

	// Empty frontend operations
	c.Assert(s.Engine.UpsertFrontend(engine.Frontend{}, 0), FitsTypeOf, &engine.InvalidFormatError{})
	c.Assert(s.Engine.DeleteFrontend(engine.FrontendKey{}), FitsTypeOf, &engine.InvalidFormatError{})
}

func (s *EngineSuite) HostCRUD(c *C) {
	host := engine.Host{Name: "localhost"}

	c.Assert(s.Engine.UpsertHost(host), IsNil)
	s.expectChanges(c, &engine.HostUpserted{Host: host})

	hs, err := s.Engine.GetHosts()
	c.Assert(err, IsNil)
	c.Assert(hs, DeepEquals, []engine.Host{host})

	hk := engine.HostKey{Name: "localhost"}
	c.Assert(s.Engine.DeleteHost(hk), IsNil)

	s.expectChanges(c, &engine.HostDeleted{HostKey: hk})
}

func (s *EngineSuite) HostWithKeyPair(c *C) {
	host := engine.Host{Name: "localhost"}

	host.Settings.Default = true
	host.Settings.KeyPair = &engine.KeyPair{
		Key:  []byte("hello"),
		Cert: []byte("world"),
	}

	c.Assert(s.Engine.UpsertHost(host), IsNil)
	s.expectChanges(c, &engine.HostUpserted{Host: host})

	hk := engine.HostKey{Name: host.Name}
	c.Assert(s.Engine.DeleteHost(hk), IsNil)

	s.expectChanges(c, &engine.HostDeleted{
		HostKey: hk,
	})
}

func (s *EngineSuite) HostWithOCSP(c *C) {
	host := engine.Host{Name: "localhost"}

	host.Settings.Default = true
	host.Settings.KeyPair = &engine.KeyPair{
		Key:  []byte("hello"),
		Cert: []byte("world"),
	}

	host.Settings.OCSP = engine.OCSPSettings{
		Enabled:            true,
		Responders:         []string{"http://a.com", "http://b.com"},
		SkipSignatureCheck: true,
		Period:             "1h",
	}

	c.Assert(s.Engine.UpsertHost(host), IsNil)
	s.expectChanges(c, &engine.HostUpserted{Host: host})

	hk := engine.HostKey{Name: host.Name}
	h2, err := s.Engine.GetHost(hk)
	c.Assert(err, IsNil)
	c.Assert(h2, DeepEquals, &host)
}

func (s *EngineSuite) HostUpsertKeyPair(c *C) {
	host := engine.Host{Name: "localhost"}

	c.Assert(s.Engine.UpsertHost(host), IsNil)

	hostNoKeyPair := host
	hostNoKeyPair.Settings.KeyPair = nil

	host.Settings.KeyPair = &engine.KeyPair{
		Key:  []byte("hello"),
		Cert: []byte("world"),
	}
	c.Assert(s.Engine.UpsertHost(host), IsNil)

	s.expectChanges(c,
		&engine.HostUpserted{Host: hostNoKeyPair},
		&engine.HostUpserted{Host: host})
}

func (s *EngineSuite) ListenerCRUD(c *C) {
	listener := engine.Listener{
		Id:       "l1",
		Protocol: "http",
		Address: engine.Address{
			Network: "tcp",
			Address: "127.0.0.1:9000",
		},
	}
	c.Assert(s.Engine.UpsertListener(listener), IsNil)
	lk := engine.ListenerKey{Id: listener.Id}

	out, err := s.Engine.GetListener(lk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &listener)

	ls, err := s.Engine.GetListeners()
	c.Assert(err, IsNil)
	c.Assert(ls, DeepEquals, []engine.Listener{listener})

	s.expectChanges(c,
		&engine.ListenerUpserted{Listener: listener},
	)
	c.Assert(s.Engine.DeleteListener(lk), IsNil)

	s.expectChanges(c,
		&engine.ListenerDeleted{ListenerKey: lk},
	)
}

func (s *EngineSuite) ListenerSettingsCRUD(c *C) {
	listener := engine.Listener{
		Id:       "l1",
		Protocol: "https",
		Address: engine.Address{
			Network: "tcp",
			Address: "127.0.0.1:9000",
		},
		Settings: &engine.HTTPSListenerSettings{
			TLS: engine.TLSSettings{
				InsecureSkipVerify: true,
				CipherSuites: []string{
					"TLS_RSA_WITH_AES_256_CBC_SHA",
					"TLS_RSA_WITH_AES_128_CBC_SHA",
				},
			},
		},
	}
	c.Assert(s.Engine.UpsertListener(listener), IsNil)
	lk := engine.ListenerKey{Id: listener.Id}

	out, err := s.Engine.GetListener(lk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &listener)

	ls, err := s.Engine.GetListeners()
	c.Assert(err, IsNil)
	c.Assert(ls, DeepEquals, []engine.Listener{listener})

	s.expectChanges(c,
		&engine.ListenerUpserted{Listener: listener},
	)
	c.Assert(s.Engine.DeleteListener(lk), IsNil)

	s.expectChanges(c,
		&engine.ListenerDeleted{ListenerKey: lk},
	)
}

func (s *EngineSuite) BackendCRUD(c *C) {
	b := engine.Backend{Id: "b1", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}

	c.Assert(s.Engine.UpsertBackend(b), IsNil)

	s.expectChanges(c, &engine.BackendUpserted{Backend: b})

	bk := engine.BackendKey{Id: b.Id}

	out, err := s.Engine.GetBackend(bk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &b)

	bs, err := s.Engine.GetBackends()
	c.Assert(len(bs), Equals, 1)
	c.Assert(bs[0], DeepEquals, b)

	b.Settings = engine.HTTPBackendSettings{Timeouts: engine.HTTPBackendTimeouts{Read: "1s"}}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)

	s.expectChanges(c, &engine.BackendUpserted{Backend: b})

	b.Settings = engine.HTTPBackendSettings{
		Timeouts: engine.HTTPBackendTimeouts{Read: "1s"},
		TLS:      &engine.TLSSettings{PreferServerCipherSuites: true},
	}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)

	s.expectChanges(c, &engine.BackendUpserted{Backend: b})

	err = s.Engine.DeleteBackend(bk)
	c.Assert(err, IsNil)

	s.expectChanges(c, &engine.BackendDeleted{
		BackendKey: bk,
	})
}

func (s *EngineSuite) BackendDeleteUsed(c *C) {
	b := engine.Backend{Id: "b0", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)

	f := engine.Frontend{
		Id:        "f1",
		Route:     `Path("/hello")`,
		BackendId: b.Id,
		Type:      engine.HTTP,
		Settings:  engine.HTTPFrontendSettings{},
	}
	c.Assert(s.Engine.UpsertFrontend(f, 0), IsNil)

	s.collectChanges(c, 2)

	c.Assert(s.Engine.DeleteBackend(engine.BackendKey{Id: b.Id}), NotNil)
}

func (s *EngineSuite) BackendDeleteUnused(c *C) {
	b := engine.Backend{Id: "b0", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	b1 := engine.Backend{Id: "b1", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)
	c.Assert(s.Engine.UpsertBackend(b1), IsNil)

	f := engine.Frontend{
		Id:        "f1",
		Route:     `Path("/hello")`,
		BackendId: b.Id,
		Type:      engine.HTTP,
		Settings:  engine.HTTPFrontendSettings{},
	}
	c.Assert(s.Engine.UpsertFrontend(f, 0), IsNil)

	s.collectChanges(c, 2)

	c.Assert(s.Engine.DeleteBackend(engine.BackendKey{Id: b.Id}), NotNil)
	c.Assert(s.Engine.DeleteBackend(engine.BackendKey{Id: b1.Id}), IsNil)
}

func (s *EngineSuite) ServerCRUD(c *C) {
	b := engine.Backend{Id: "b0", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}

	c.Assert(s.Engine.UpsertBackend(b), IsNil)

	s.expectChanges(c, &engine.BackendUpserted{Backend: b})

	srv := engine.Server{Id: "srv0", URL: "http://localhost:1000"}
	bk := engine.BackendKey{Id: b.Id}
	sk := engine.ServerKey{BackendKey: bk, Id: srv.Id}

	c.Assert(s.Engine.UpsertServer(bk, srv, 0), IsNil)

	srvo, err := s.Engine.GetServer(sk)
	c.Assert(err, IsNil)
	c.Assert(srvo, DeepEquals, &srv)

	srvs, err := s.Engine.GetServers(bk)
	c.Assert(err, IsNil)
	c.Assert(srvs, DeepEquals, []engine.Server{srv})

	s.expectChanges(c, &engine.ServerUpserted{
		BackendKey: bk,
		Server:     srv,
	})

	err = s.Engine.DeleteServer(sk)
	c.Assert(err, IsNil)

	s.expectChanges(c, &engine.ServerDeleted{
		ServerKey: sk,
	})
}

func (s *EngineSuite) ServerExpire(c *C) {
	b := engine.Backend{Id: "b0", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}

	c.Assert(s.Engine.UpsertBackend(b), IsNil)
	s.collectChanges(c, 1)

	srv := engine.Server{Id: "srv0", URL: "http://localhost:1000"}
	bk := engine.BackendKey{Id: b.Id}
	sk := engine.ServerKey{BackendKey: bk, Id: srv.Id}
	c.Assert(s.Engine.UpsertServer(bk, srv, time.Second), IsNil)

	s.expectChanges(c,
		&engine.ServerUpserted{
			BackendKey: bk,
			Server:     srv,
		}, &engine.ServerDeleted{
			ServerKey: sk,
		})
}

func (s *EngineSuite) FrontendCRUD(c *C) {
	b := engine.Backend{Id: "b0", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)
	s.collectChanges(c, 1)

	f := engine.Frontend{
		Id:        "f1",
		BackendId: b.Id,
		Route:     `Path("/hello")`,
		Type:      engine.HTTP,
		Settings:  engine.HTTPFrontendSettings{},
	}

	c.Assert(s.Engine.UpsertFrontend(f, 0), IsNil)

	fk := engine.FrontendKey{Id: f.Id}
	out, err := s.Engine.GetFrontend(fk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &f)

	s.expectChanges(c, &engine.FrontendUpserted{
		Frontend: f,
	})

	// Make some updates
	b1 := engine.Backend{Id: "b1", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	c.Assert(s.Engine.UpsertBackend(b1), IsNil)
	s.collectChanges(c, 1)

	f.BackendId = "b1"
	f.Settings = engine.HTTPFrontendSettings{
		Hostname: "host1",
	}
	c.Assert(s.Engine.UpsertFrontend(f, 0), IsNil)

	out, err = s.Engine.GetFrontend(fk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &f)

	s.expectChanges(c, &engine.FrontendUpserted{
		Frontend: f,
	})

	// Delete
	c.Assert(s.Engine.DeleteFrontend(fk), IsNil)
	s.expectChanges(c, &engine.FrontendDeleted{
		FrontendKey: fk,
	})
}

func (s *EngineSuite) FrontendExpire(c *C) {
	b := engine.Backend{Id: "b0", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)
	s.collectChanges(c, 1)

	f := engine.Frontend{
		Id:        "f1",
		BackendId: b.Id,
		Route:     `Path("/hello")`,
		Type:      engine.HTTP,
		Settings:  engine.HTTPFrontendSettings{},
	}
	c.Assert(s.Engine.UpsertFrontend(f, time.Second), IsNil)

	fk := engine.FrontendKey{Id: f.Id}
	s.expectChanges(c,
		&engine.FrontendUpserted{
			Frontend: f,
		}, &engine.FrontendDeleted{
			FrontendKey: fk,
		})
}

func (s *EngineSuite) FrontendBadBackend(c *C) {
	c.Assert(
		s.Engine.UpsertFrontend(engine.Frontend{
			Id:        "f1",
			Type:      engine.HTTP,
			BackendId: "Nonexistent",
			Route:     `Path("/hello")`,
			Settings:  engine.HTTPFrontendSettings{},
		}, 0),
		NotNil)
}

func (s *EngineSuite) MiddlewareCRUD(c *C) {
	b := engine.Backend{Id: "b1", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)

	f := engine.Frontend{
		Id:        "f1",
		Type:      engine.HTTP,
		Route:     `Path("/hello")`,
		Settings:  engine.HTTPFrontendSettings{},
		BackendId: b.Id,
	}
	c.Assert(s.Engine.UpsertFrontend(f, 0), IsNil)
	s.collectChanges(c, 2)

	fk := engine.FrontendKey{Id: f.Id}
	m := s.makeConnLimit("cl1", "client.ip", 10)
	c.Assert(s.Engine.UpsertMiddleware(fk, m, 0), IsNil)
	s.expectChanges(c, &engine.MiddlewareUpserted{
		FrontendKey: fk,
		Middleware:  m,
	})

	mk := engine.MiddlewareKey{Id: m.Id, FrontendKey: fk}
	out, err := s.Engine.GetMiddleware(mk)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, &m)

	// Let us upsert middleware
	m.Middleware.(*connlimit.ConnLimit).Connections = 100
	c.Assert(s.Engine.UpsertMiddleware(fk, m, 0), IsNil)
	s.expectChanges(c, &engine.MiddlewareUpserted{
		FrontendKey: fk,
		Middleware:  m,
	})

	ms, err := s.Engine.GetMiddlewares(fk)
	c.Assert(err, IsNil)
	c.Assert(ms, DeepEquals, []engine.Middleware{m})

	c.Assert(s.Engine.DeleteMiddleware(mk), IsNil)

	s.expectChanges(c, &engine.MiddlewareDeleted{
		MiddlewareKey: mk,
	})
}

func (s *EngineSuite) MiddlewareExpire(c *C) {
	b := engine.Backend{Id: "b1", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)

	f := engine.Frontend{
		Id:        "f1",
		Route:     `Path("/hello")`,
		BackendId: b.Id,
		Type:      engine.HTTP,
		Settings:  engine.HTTPFrontendSettings{},
	}
	c.Assert(s.Engine.UpsertFrontend(f, 0), IsNil)
	s.collectChanges(c, 2)

	fk := engine.FrontendKey{Id: f.Id}
	m := s.makeConnLimit("cl1", "client.ip", 10)
	c.Assert(s.Engine.UpsertMiddleware(fk, m, time.Second), IsNil)
	s.expectChanges(c, &engine.MiddlewareUpserted{
		FrontendKey: fk,
		Middleware:  m,
	})

	mk := engine.MiddlewareKey{Id: m.Id, FrontendKey: fk}
	s.expectChanges(c, &engine.MiddlewareDeleted{
		MiddlewareKey: mk,
	})
}

func (s *EngineSuite) MiddlewareBadFrontend(c *C) {
	fk := engine.FrontendKey{Id: "wrong"}
	m := s.makeConnLimit("cl1", "client.ip", 10)
	c.Assert(s.Engine.UpsertMiddleware(fk, m, 0), NotNil)
}

func (s *EngineSuite) MiddlewareBadType(c *C) {
	fk := engine.FrontendKey{Id: "wrong"}
	m := s.makeConnLimit("cl1", "client.ip", 10)

	b := engine.Backend{Id: "b1", Type: engine.HTTP, Settings: engine.HTTPBackendSettings{}}
	c.Assert(s.Engine.UpsertBackend(b), IsNil)

	f := engine.Frontend{
		Id:        "f1",
		Route:     `Path("/hello")`,
		Type:      engine.HTTP,
		BackendId: b.Id,
		Settings:  engine.HTTPFrontendSettings{},
	}
	c.Assert(s.Engine.UpsertFrontend(f, 0), IsNil)
	s.collectChanges(c, 2)
	m.Type = "blabla"
	c.Assert(s.Engine.UpsertMiddleware(fk, m, 0), NotNil)
}
