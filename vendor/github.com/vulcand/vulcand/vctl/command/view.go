package command

import (
	"fmt"
	"sort"

	"github.com/buger/goterm"
	"github.com/vulcand/vulcand/engine"
)

func hostsView(hs []engine.Host) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Name\tDefault\n")

	if len(hs) == 0 {
		return t.String()
	}
	for _, h := range hs {
		fmt.Fprint(t, hostView(&h))
	}
	return t.String()
}

func hostView(h *engine.Host) string {
	return fmt.Sprintf("%s\t%t\n", h.Name, h.Settings.Default)
}

func listenersView(ls []engine.Listener) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Id\tProtocol\tNetwork\tAddress\tScope\n")

	if len(ls) == 0 {
		return t.String()
	}
	for _, l := range ls {
		fmt.Fprint(t, listenerView(&l))
	}
	return t.String()
}

func listenerView(l *engine.Listener) string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", l.Id, l.Protocol, l.Address.Network, l.Address.Address, l.Scope)
}

func frontendsView(fs []engine.Frontend) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Id\tRoute\tBackend\tType\n")

	if len(fs) == 0 {
		return t.String()
	}
	for _, v := range fs {
		fmt.Fprint(t, frontendView(&v))
	}
	return t.String()
}

func frontendView(f *engine.Frontend) string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\n", f.Id, f.Route, f.BackendId, f.Type)
}

func backendsView(bs []engine.Backend) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Id\tType\n")

	if len(bs) == 0 {
		return t.String()
	}
	for _, v := range bs {
		fmt.Fprint(t, backendView(&v))
	}
	return t.String()
}

func backendView(b *engine.Backend) string {
	return fmt.Sprintf("%s\t%s\n", b.Id, b.Type)
}

func serversView(srvs []engine.Server) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Id\tURL\n")
	if len(srvs) == 0 {
		return t.String()
	}
	for _, v := range srvs {
		fmt.Fprint(t, serverView(&v))
	}
	return t.String()
}

func serverView(s *engine.Server) string {
	return fmt.Sprintf("%s\t%s\n", s.Id, s.URL)
}

func middlewaresView(ms []engine.Middleware) string {
	sort.Sort(&middlewareSorter{ms: ms})

	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Id\tPriority\tType\tSettings\n")
	if len(ms) == 0 {
		return t.String()
	}
	for _, v := range ms {
		fmt.Fprint(t, middlewareView(&v))
	}
	return t.String()
}

func middlewareView(m *engine.Middleware) string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\n", m.Id, m.Priority, m.Type, m.Middleware)
}

// Sorts middlewares by their priority
type middlewareSorter struct {
	ms []engine.Middleware
}

// Len is part of sort.Interface.
func (s *middlewareSorter) Len() int {
	return len(s.ms)
}

// Swap is part of sort.Interface.
func (s *middlewareSorter) Swap(i, j int) {
	s.ms[i], s.ms[j] = s.ms[j], s.ms[i]
}

func (s *middlewareSorter) Less(i, j int) bool {
	return s.ms[i].Priority < s.ms[j].Priority
}
