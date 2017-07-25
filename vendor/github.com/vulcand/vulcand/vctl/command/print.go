package command

import (
	"bytes"
	"fmt"
	"io"

	"github.com/buger/goterm"
	"github.com/vulcand/vulcand/engine"
)

func (cmd *Command) printResult(format string, in interface{}, err error) {
	if err != nil {
		cmd.printError(err)
	} else {
		cmd.printOk(format, fmt.Sprintf("%v", in))
	}
}

func (cmd *Command) printStatus(in interface{}, err error) {
	if err != nil {
		cmd.printError(err)
	} else {
		cmd.printOk("%s", in)
	}
}

func (cmd *Command) printError(err error) {
	fmt.Fprint(cmd.out, goterm.Color(fmt.Sprintf("ERROR: %s", err), goterm.RED)+"\n")
}

func (cmd *Command) printOk(message string, params ...interface{}) {
	fmt.Fprintf(cmd.out, goterm.Color(fmt.Sprintf("OK: %s\n", fmt.Sprintf(message, params...)), goterm.GREEN)+"\n")
}

func (cmd *Command) printInfo(message string, params ...interface{}) {
	fmt.Fprintf(cmd.out, "INFO: %s\n", fmt.Sprintf(message, params...))
}

func (cmd *Command) printHosts(hosts []engine.Host) {
	fmt.Fprintf(cmd.out, "\n[Hosts]\n")
	writeS(cmd.out, hostsView(hosts))
}

func (cmd *Command) printHost(host *engine.Host) {
	fmt.Fprintf(cmd.out, "\n[Host]\n")
	writeS(cmd.out, hostsView([]engine.Host{*host}))
}

func (cmd *Command) printListeners(ls []engine.Listener) {
	fmt.Fprintf(cmd.out, "\n[Listeners]\n")
	writeS(cmd.out, listenersView(ls))
}

func (cmd *Command) printListener(l *engine.Listener) {
	fmt.Fprintf(cmd.out, "\n[Listeners]\n")
	writeS(cmd.out, listenersView([]engine.Listener{*l}))
}

func (cmd *Command) printServers(srvs []engine.Server) {
	fmt.Fprintf(cmd.out, "\n[Servers]\n")
	writeS(cmd.out, serversView(srvs))
}

func (cmd *Command) printServer(s *engine.Server) {
	fmt.Fprintf(cmd.out, "\n[Server]\n")
	writeS(cmd.out, serversView([]engine.Server{*s}))
}

func (cmd *Command) printOverview(frontend []engine.Frontend, servers []engine.Server) {
	out := &bytes.Buffer{}
	fmt.Fprintf(out, "\n[Frontend]\n")
	out.WriteString(frontendsOverview(frontend))
	fmt.Fprintf(cmd.out, out.String())

	out = &bytes.Buffer{}
	fmt.Fprintf(out, "\n[Servers]\n")
	out.WriteString(serversOverview(servers))
	fmt.Fprintf(cmd.out, out.String())
}

func (cmd *Command) printBackends(bs []engine.Backend) {
	fmt.Fprintf(cmd.out, "\n[Backends]\n")
	writeS(cmd.out, backendsView(bs))
}

func (cmd *Command) printBackend(b *engine.Backend, srvs []engine.Server) {
	fmt.Fprintf(cmd.out, "\n[Backend]\n")
	writeS(cmd.out, backendsView([]engine.Backend{*b}))
	fmt.Fprintf(cmd.out, "\n[Servers]\n")
	writeS(cmd.out, serversView(srvs))
}

func (cmd *Command) printFrontends(fs []engine.Frontend) {
	fmt.Fprintf(cmd.out, "\n[Frontends]\n")
	writeS(cmd.out, frontendsView(fs))
}

func (cmd *Command) printFrontend(f *engine.Frontend, ms []engine.Middleware) {
	fmt.Fprintf(cmd.out, "\n[Frontend]\n")
	writeS(cmd.out, frontendsView([]engine.Frontend{*f}))
	fmt.Fprintf(cmd.out, "\n[Middlewares]\n")
	writeS(cmd.out, middlewaresView(ms))
}

func writeS(w io.Writer, v string) {
	w.Write([]byte(v))
}
