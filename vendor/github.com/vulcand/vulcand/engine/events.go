package engine

import "fmt"

type HostUpserted struct {
	Host Host
}

func (h *HostUpserted) String() string {
	return fmt.Sprintf("HostUpserted(host=%v)", &h.Host)
}

type HostDeleted struct {
	HostKey HostKey
}

func (h *HostDeleted) String() string {
	return fmt.Sprintf("HostDeleted(hostKey=%v)", &h.HostKey)
}

type ListenerUpserted struct {
	HostKey  HostKey
	Listener Listener
}

func (l *ListenerUpserted) String() string {
	return fmt.Sprintf("ListenerUpserted(hostKey=%v, l=%v)", &l.HostKey, &l.Listener)
}

type ListenerDeleted struct {
	ListenerKey ListenerKey
}

func (l *ListenerDeleted) String() string {
	return fmt.Sprintf("ListenerDeleted(listenerKey=%v)", &l.ListenerKey)
}

type FrontendUpserted struct {
	Frontend Frontend
}

func (f *FrontendUpserted) String() string {
	return fmt.Sprintf("FrontendUpserted(frontend=%v)", &f.Frontend)
}

type FrontendDeleted struct {
	FrontendKey FrontendKey
}

func (f *FrontendDeleted) String() string {
	return fmt.Sprintf("FrontendDeleted(frontendKey=%v)", &f.FrontendKey)
}

type MiddlewareUpserted struct {
	FrontendKey FrontendKey
	Middleware  Middleware
}

func (m *MiddlewareUpserted) String() string {
	return fmt.Sprintf("MiddlewareUpserted(frontendKey=%v, middleware=%v)", &m.FrontendKey, &m.Middleware)
}

type MiddlewareDeleted struct {
	MiddlewareKey MiddlewareKey
}

func (m *MiddlewareDeleted) String() string {
	return fmt.Sprintf("MiddlewareDeleted(middlewareKey=%v)", &m.MiddlewareKey)
}

type BackendUpserted struct {
	Backend Backend
}

func (b *BackendUpserted) String() string {
	return fmt.Sprintf("BackendUpserted(backend=%v)", &b.Backend)
}

type BackendDeleted struct {
	BackendKey BackendKey
}

func (b *BackendDeleted) String() string {
	return fmt.Sprintf("BackendDeleted(backendKey=%v)", &b.BackendKey)
}

type ServerUpserted struct {
	BackendKey BackendKey
	Server     Server
}

func (s *ServerUpserted) String() string {
	return fmt.Sprintf("ServerUpserted(backendKey=%v, server=%v)", &s.BackendKey, &s.Server)
}

type ServerDeleted struct {
	ServerKey ServerKey
}

func (s *ServerDeleted) String() string {
	return fmt.Sprintf("ServerDeleted(serverKey=%v)", &s.ServerKey)
}
