// package etcdng contains the implementation of the Etcd-backed engine, where all vulcand properties are implemented as directories or keys.
// this engine is capable of watching the changes and generating events.
package etcdv2ng

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"crypto/x509"
	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/plugin"
	"github.com/vulcand/vulcand/secret"
	"golang.org/x/net/context"
	"io/ioutil"
)

type ng struct {
	nodes            []string
	registry         *plugin.Registry
	etcdKey          string
	client           etcd.Client
	kapi             etcd.KeysAPI
	context          context.Context
	cancelFunc       context.CancelFunc
	cancelC          chan bool
	stopC            chan bool
	syncClusterStopC chan bool
	logsev           log.Level
	options          Options
	requireQuorum    bool
}

type Options struct {
	EtcdConsistency         string
	EtcdCaFile              string
	EtcdCertFile            string
	EtcdKeyFile             string
	EtcdSyncIntervalSeconds int64
	Box                     *secret.Box
}

func New(nodes []string, etcdKey string, registry *plugin.Registry, options Options) (engine.Engine, error) {
	n := &ng{
		nodes:            nodes,
		registry:         registry,
		etcdKey:          etcdKey,
		cancelC:          make(chan bool, 1),
		stopC:            make(chan bool, 1),
		syncClusterStopC: make(chan bool, 1),
		options:          options,
	}
	if err := n.reconnect(); err != nil {
		return nil, err
	}
	if options.EtcdSyncIntervalSeconds > 0 {
		go n.client.AutoSync(n.context, time.Duration(n.options.EtcdSyncIntervalSeconds)*time.Second)
	}
	return n, nil
}

func (n *ng) Close() {
	n.syncClusterStopC <- true
	if n.cancelFunc != nil {
		n.cancelFunc()
	}
}

func (n *ng) GetLogSeverity() log.Level {
	return n.logsev
}

func (n *ng) SetLogSeverity(sev log.Level) {
	n.logsev = sev
	log.SetLevel(n.logsev)
}

func (n *ng) reconnect() error {
	n.Close()
	var client etcd.Client
	cfg := n.getEtcdClientConfig()
	var err error
	if client, err = etcd.New(cfg); err != nil {
		return err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	n.context = ctx
	n.cancelFunc = cancelFunc
	n.client = client
	n.kapi = etcd.NewKeysAPI(n.client)
	n.requireQuorum = true
	if n.options.EtcdConsistency == "WEAK" {
		n.requireQuorum = false
	}
	return nil
}

func (n *ng) getEtcdClientConfig() etcd.Config {
	return etcd.Config{
		Endpoints: n.nodes,
		Transport: n.newHttpTransport(),
	}
}

func (n *ng) newHttpTransport() etcd.CancelableTransport {

	var cc *tls.Config = nil

	if n.options.EtcdCertFile != "" && n.options.EtcdKeyFile != "" {
		var rpool *x509.CertPool = nil
		if n.options.EtcdCaFile != "" {
			if pemBytes, err := ioutil.ReadFile(n.options.EtcdCaFile); err == nil {
				rpool = x509.NewCertPool()
				rpool.AppendCertsFromPEM(pemBytes)
			} else {
				log.Errorf("Error reading Etcd Cert CA File: %v", err)
			}
		}

		if tlsCert, err := tls.LoadX509KeyPair(n.options.EtcdCertFile, n.options.EtcdKeyFile); err == nil {
			cc = &tls.Config{
				RootCAs:            rpool,
				Certificates:       []tls.Certificate{tlsCert},
				InsecureSkipVerify: true,
			}
		} else {
			log.Errorf("Error loading KeyPair for TLS client: %v", err)
		}

	}

	//Copied from etcd.DefaultTransport declaration
	//Wasn't sure how to make a clean reliable deep-copy, and instead
	//creating a new object was the safest and most reliable assurance
	//that we aren't overwriting some global struct potentially
	//shared by other etcd users.
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     cc,
	}

	return tr
}

func (n *ng) GetRegistry() *plugin.Registry {
	return n.registry
}

func (n *ng) GetHosts() ([]engine.Host, error) {
	hosts := []engine.Host{}
	vals, err := n.getDirs(n.etcdKey, "hosts")
	if err != nil {
		return nil, err
	}
	for _, hostKey := range vals {
		host, err := n.GetHost(engine.HostKey{suffix(hostKey)})
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, *host)
	}
	return hosts, nil
}

func (n *ng) GetHost(key engine.HostKey) (*engine.Host, error) {
	hostKey := n.path("hosts", key.Name, "host")

	var host *host
	err := n.getJSONVal(hostKey, &host)
	if err != nil {
		return nil, err
	}

	var keyPair *engine.KeyPair
	if len(host.Settings.KeyPair) != 0 {
		if err := n.openSealedJSONVal(host.Settings.KeyPair, &keyPair); err != nil {
			return nil, err
		}
	}

	return engine.NewHost(key.Name, engine.HostSettings{Default: host.Settings.Default, KeyPair: keyPair, OCSP: host.Settings.OCSP})
}

func (n *ng) UpsertHost(h engine.Host) error {
	if h.Name == "" {
		return &engine.InvalidFormatError{Message: "hostname can not be empty"}
	}
	hostKey := n.path("hosts", h.Name, "host")

	val := host{
		Name: h.Name,
		Settings: hostSettings{
			Default: h.Settings.Default,
			OCSP:    h.Settings.OCSP,
		},
	}

	if h.Settings.KeyPair != nil {
		bytes, err := n.sealJSONVal(h.Settings.KeyPair)
		if err != nil {
			return err
		}
		val.Settings.KeyPair = bytes
	}

	return n.setJSONVal(hostKey, val, noTTL)
}

func (n *ng) DeleteHost(key engine.HostKey) error {
	if key.Name == "" {
		return &engine.InvalidFormatError{Message: "hostname can not be empty"}
	}
	return n.deleteKey(n.path("hosts", key.Name))
}

func (n *ng) GetListeners() ([]engine.Listener, error) {
	ls := []engine.Listener{}
	vals, err := n.getVals(n.etcdKey, "listeners")
	if err != nil {
		return nil, err
	}
	for _, p := range vals {
		l, err := n.GetListener(engine.ListenerKey{Id: suffix(p.Key)})
		if err != nil {
			return nil, err
		}
		ls = append(ls, *l)
	}
	return ls, nil
}

func (n *ng) GetListener(key engine.ListenerKey) (*engine.Listener, error) {
	bytes, err := n.getVal(n.path("listeners", key.Id))
	if err != nil {
		return nil, err
	}
	l, err := engine.ListenerFromJSON([]byte(bytes), key.Id)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (n *ng) UpsertListener(listener engine.Listener) error {
	if listener.Id == "" {
		return &engine.InvalidFormatError{Message: "listener id can not be empty"}
	}
	return n.setJSONVal(n.path("listeners", listener.Id), listener, noTTL)
}

func (s *ng) DeleteListener(key engine.ListenerKey) error {
	if key.Id == "" {
		return &engine.InvalidFormatError{Message: "listener id can not be empty"}
	}
	return s.deleteKey(s.path("listeners", key.Id))
}

func (n *ng) UpsertFrontend(f engine.Frontend, ttl time.Duration) error {
	if f.Id == "" {
		return &engine.InvalidFormatError{Message: "frontend id can not be empty"}
	}
	if _, err := n.GetBackend(engine.BackendKey{Id: f.BackendId}); err != nil {
		return err
	}
	if err := n.setJSONVal(n.path("frontends", f.Id, "frontend"), f, noTTL); err != nil {
		return err
	}
	if ttl == 0 {
		return nil
	}
	_, err := n.kapi.Set(n.context, n.path("frontends", f.Id), "", &etcd.SetOptions{Dir: true, TTL: ttl})
	return convertErr(err)
}

func (n *ng) GetFrontends() ([]engine.Frontend, error) {
	fs := []engine.Frontend{}
	vals, err := n.getDirs(n.etcdKey, "frontends")
	if err != nil {
		return nil, err
	}
	for _, fPath := range vals {
		f, err := n.GetFrontend(engine.FrontendKey{suffix(fPath)})
		if err != nil {
			return nil, err
		}
		fs = append(fs, *f)
	}
	return fs, nil
}

func (n *ng) GetFrontend(key engine.FrontendKey) (*engine.Frontend, error) {
	frontendKey := n.path("frontends", key.Id, "frontend")

	bytes, err := n.getVal(frontendKey)
	if err != nil {
		return nil, err
	}
	return engine.FrontendFromJSON(n.registry.GetRouter(), []byte(bytes), key.Id)
}

func (n *ng) DeleteFrontend(fk engine.FrontendKey) error {
	if fk.Id == "" {
		return &engine.InvalidFormatError{Message: "frontend id can not be empty"}
	}
	return n.deleteKey(n.path("frontends", fk.Id))
}

func (n *ng) GetBackends() ([]engine.Backend, error) {
	backends := []engine.Backend{}
	ups, err := n.getDirs(n.etcdKey, "backends")
	if err != nil {
		return nil, err
	}
	for _, backendKey := range ups {
		b, err := n.GetBackend(engine.BackendKey{Id: suffix(backendKey)})
		if err != nil {
			return nil, err
		}
		backends = append(backends, *b)
	}
	return backends, nil
}

func (n *ng) GetBackend(key engine.BackendKey) (*engine.Backend, error) {
	backendKey := n.path("backends", key.Id, "backend")

	bytes, err := n.getVal(backendKey)
	if err != nil {
		return nil, err
	}
	return engine.BackendFromJSON([]byte(bytes), key.Id)
}

func (n *ng) UpsertBackend(b engine.Backend) error {
	if b.Id == "" {
		return &engine.InvalidFormatError{Message: "backend id can not be empty"}
	}
	return n.setJSONVal(n.path("backends", b.Id, "backend"), b, noTTL)
}

func (n *ng) DeleteBackend(bk engine.BackendKey) error {
	if bk.Id == "" {
		return &engine.InvalidFormatError{Message: "backend id can not be empty"}
	}
	fs, err := n.backendUsedBy(bk)
	if err != nil {
		return err
	}
	if len(fs) != 0 {
		return fmt.Errorf("can not delete backend '%v', it is in use by %s", bk, fs)
	}
	_, err = n.kapi.Delete(n.context, n.path("backends", bk.Id), &etcd.DeleteOptions{Recursive: true})
	return convertErr(err)
}

func (n *ng) GetMiddlewares(fk engine.FrontendKey) ([]engine.Middleware, error) {
	ms := []engine.Middleware{}
	keys, err := n.getVals(n.etcdKey, "frontends", fk.Id, "middlewares")
	if err != nil {
		return nil, err
	}
	for _, p := range keys {
		m, err := n.GetMiddleware(engine.MiddlewareKey{Id: suffix(p.Key), FrontendKey: fk})
		if err != nil {
			return nil, err
		}
		ms = append(ms, *m)
	}
	return ms, nil
}

func (n *ng) GetMiddleware(key engine.MiddlewareKey) (*engine.Middleware, error) {
	mKey := n.path("frontends", key.FrontendKey.Id, "middlewares", key.Id)
	bytes, err := n.getVal(mKey)
	if err != nil {
		return nil, err
	}
	return engine.MiddlewareFromJSON([]byte(bytes), n.registry.GetSpec, key.Id)
}

func (n *ng) UpsertMiddleware(fk engine.FrontendKey, m engine.Middleware, ttl time.Duration) error {
	if fk.Id == "" || m.Id == "" {
		return &engine.InvalidFormatError{Message: "frontend id and middleware id can not be empty"}
	}
	if _, err := n.GetFrontend(fk); err != nil {
		return err
	}
	return n.setJSONVal(n.path("frontends", fk.Id, "middlewares", m.Id), m, ttl)
}

func (n *ng) DeleteMiddleware(mk engine.MiddlewareKey) error {
	if mk.FrontendKey.Id == "" || mk.Id == "" {
		return &engine.InvalidFormatError{Message: "frontend id and middleware id can not be empty"}
	}
	return n.deleteKey(n.path("frontends", mk.FrontendKey.Id, "middlewares", mk.Id))
}

func (n *ng) UpsertServer(bk engine.BackendKey, s engine.Server, ttl time.Duration) error {
	if s.Id == "" || bk.Id == "" {
		return &engine.InvalidFormatError{Message: "backend id and server id can not be empty"}
	}
	if _, err := n.GetBackend(bk); err != nil {
		return err
	}
	return n.setJSONVal(n.path("backends", bk.Id, "servers", s.Id), s, ttl)
}

func (n *ng) GetServers(bk engine.BackendKey) ([]engine.Server, error) {
	svs := []engine.Server{}
	keys, err := n.getVals(n.etcdKey, "backends", bk.Id, "servers")
	if err != nil {
		return nil, err
	}
	for _, p := range keys {
		srv, err := n.GetServer(engine.ServerKey{Id: suffix(p.Key), BackendKey: bk})
		if err != nil {
			return nil, err
		}
		svs = append(svs, *srv)
	}
	return svs, nil
}

func (n *ng) GetServer(sk engine.ServerKey) (*engine.Server, error) {
	bytes, err := n.getVal(n.path("backends", sk.BackendKey.Id, "servers", sk.Id))
	if err != nil {
		return nil, err
	}
	return engine.ServerFromJSON([]byte(bytes), sk.Id)
}

func (n *ng) DeleteServer(sk engine.ServerKey) error {
	if sk.Id == "" || sk.BackendKey.Id == "" {
		return &engine.InvalidFormatError{Message: "backend id and server id can not be empty"}
	}
	return n.deleteKey(n.path("backends", sk.BackendKey.Id, "servers", sk.Id))
}

func (n *ng) openSealedJSONVal(bytes []byte, val interface{}) error {
	if n.options.Box == nil {
		return fmt.Errorf("need secretbox to open sealed data")
	}
	sv, err := secret.SealedValueFromJSON([]byte(bytes))
	if err != nil {
		return err
	}
	unsealed, err := n.options.Box.Open(sv)
	if err != nil {
		return err
	}
	return json.Unmarshal(unsealed, val)
}

func (n *ng) sealJSONVal(val interface{}) ([]byte, error) {
	if n.options.Box == nil {
		return nil, fmt.Errorf("this backend does not support encryption")
	}
	bytes, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	v, err := n.options.Box.Seal(bytes)
	if err != nil {
		return nil, err
	}
	return secret.SealedValueToJSON(v)
}

func (n *ng) backendUsedBy(bk engine.BackendKey) ([]engine.Frontend, error) {
	fs, err := n.GetFrontends()
	usedFs := []engine.Frontend{}
	if err != nil {
		return nil, err
	}
	for _, f := range fs {
		if f.BackendId == bk.Id {
			usedFs = append(usedFs, f)
		}
	}
	return usedFs, nil
}

// Subscribe watches etcd changes and generates structured events telling vulcand to add or delete frontends, hosts etc.
// It is a blocking function.
func (n *ng) Subscribe(changes chan interface{}, cancelC chan bool) error {
	w := n.kapi.Watcher(n.etcdKey, &etcd.WatcherOptions{AfterIndex: 0, Recursive: true})
	for {
		response, err := w.Next(n.context)
		if err != nil {
			switch err {
			case context.Canceled:
				log.Infof("Stop watching: graceful shutdown")
				return nil
			default:
				log.Errorf("unexpected error: %s, stop watching", err)
				return err
			}
		}
		log.Infof("%s", responseToString(response))
		change, err := n.parseChange(response)
		if err != nil {
			log.Warningf("Ignore '%s', error: %s", responseToString(response), err)
			continue
		}
		if change != nil {
			log.Infof("%v", change)
			select {
			case changes <- change:
			case <-cancelC:
				return nil
			}
		}
	}
}

type MatcherFn func(*etcd.Response) (interface{}, error)

// Dispatches etcd key changes changes to the etcd to the matching functions
func (s *ng) parseChange(response *etcd.Response) (interface{}, error) {
	matchers := []MatcherFn{
		// Host updates
		s.parseHostChange,

		// Listener updates
		s.parseListenerChange,

		// Frontend updates
		s.parseFrontendChange,
		s.parseFrontendMiddlewareChange,

		// Backend updates
		s.parseBackendChange,
		s.parseBackendServerChange,
	}
	for _, matcher := range matchers {
		a, err := matcher(response)
		if a != nil || err != nil {
			return a, err
		}
	}
	return nil, nil
}

func (n *ng) parseHostChange(r *etcd.Response) (interface{}, error) {
	out := regexp.MustCompile("/hosts/([^/]+)(?:/host)?$").FindStringSubmatch(r.Node.Key)
	if len(out) != 2 {
		return nil, nil
	}

	hostname := out[1]

	switch r.Action {
	case createA, setA:
		host, err := n.GetHost(engine.HostKey{hostname})
		if err != nil {
			return nil, err
		}
		return &engine.HostUpserted{
			Host: *host,
		}, nil
	case deleteA, expireA:
		return &engine.HostDeleted{
			HostKey: engine.HostKey{hostname},
		}, nil
	}
	return nil, fmt.Errorf("unsupported action for host: %s", r.Action)
}

func (n *ng) parseListenerChange(r *etcd.Response) (interface{}, error) {
	out := regexp.MustCompile("/listeners/([^/]+)").FindStringSubmatch(r.Node.Key)
	if len(out) != 2 {
		return nil, nil
	}

	key := engine.ListenerKey{Id: out[1]}

	switch r.Action {
	case createA, setA:
		l, err := n.GetListener(key)
		if err != nil {
			return nil, err
		}
		return &engine.ListenerUpserted{
			Listener: *l,
		}, nil
	case deleteA, expireA:
		return &engine.ListenerDeleted{
			ListenerKey: key,
		}, nil
	}
	return nil, fmt.Errorf("unsupported action on the listener: %s", r.Action)
}

func (n *ng) parseFrontendChange(r *etcd.Response) (interface{}, error) {
	out := regexp.MustCompile("/frontends/([^/]+)(?:/frontend)?$").FindStringSubmatch(r.Node.Key)
	if len(out) != 2 {
		return nil, nil
	}
	key := engine.FrontendKey{Id: out[1]}
	switch r.Action {
	case createA, setA:
		f, err := n.GetFrontend(key)
		if err != nil {
			return nil, err
		}
		return &engine.FrontendUpserted{
			Frontend: *f,
		}, nil
	case deleteA, expireA:
		return &engine.FrontendDeleted{
			FrontendKey: key,
		}, nil
	case updateA: // this happens when we set TTL on a dir, ignore as there's no action needed from us
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported action on the frontend: %v %v", r.Node.Key, r.Action)
}

func (s *ng) parseFrontendMiddlewareChange(r *etcd.Response) (interface{}, error) {
	out := regexp.MustCompile("/frontends/([^/]+)/middlewares/([^/]+)$").FindStringSubmatch(r.Node.Key)
	if len(out) != 3 {
		return nil, nil
	}

	fk := engine.FrontendKey{Id: out[1]}
	mk := engine.MiddlewareKey{FrontendKey: fk, Id: out[2]}

	switch r.Action {
	case createA, setA:
		m, err := s.GetMiddleware(mk)
		if err != nil {
			return nil, err
		}
		return &engine.MiddlewareUpserted{
			FrontendKey: fk,
			Middleware:  *m,
		}, nil
	case deleteA, expireA:
		return &engine.MiddlewareDeleted{
			MiddlewareKey: mk,
		}, nil
	}
	return nil, fmt.Errorf("unsupported action on the rate: %s", r.Action)
}

func (n *ng) parseBackendChange(r *etcd.Response) (interface{}, error) {
	out := regexp.MustCompile("/backends/([^/]+)(?:/backend)?$").FindStringSubmatch(r.Node.Key)
	if len(out) != 2 {
		return nil, nil
	}
	bk := engine.BackendKey{Id: out[1]}
	switch r.Action {
	case createA, setA:
		b, err := n.GetBackend(bk)
		if err != nil {
			return nil, err
		}
		return &engine.BackendUpserted{
			Backend: *b,
		}, nil
	case deleteA, expireA:
		return &engine.BackendDeleted{
			BackendKey: bk,
		}, nil
	}
	return nil, fmt.Errorf("unsupported node action: %s", r.Action)
}

func (n *ng) parseBackendServerChange(r *etcd.Response) (interface{}, error) {
	out := regexp.MustCompile("/backends/([^/]+)/servers/([^/]+)$").FindStringSubmatch(r.Node.Key)
	if len(out) != 3 {
		return nil, nil
	}

	sk := engine.ServerKey{BackendKey: engine.BackendKey{Id: out[1]}, Id: out[2]}

	switch r.Action {
	case setA, createA:
		srv, err := n.GetServer(sk)
		if err != nil {
			return nil, err
		}
		return &engine.ServerUpserted{
			BackendKey: sk.BackendKey,
			Server:     *srv,
		}, nil
	case deleteA, expireA:
		return &engine.ServerDeleted{
			ServerKey: sk,
		}, nil
	case cswapA: // ignore compare and swap attempts
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported action on the server: %s", r.Action)
}

func (n ng) path(keys ...string) string {
	return strings.Join(append([]string{n.etcdKey}, keys...), "/")
}

func (n *ng) setJSONVal(key string, v interface{}, ttl time.Duration) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return n.setVal(key, bytes, ttl)
}

func (n *ng) setVal(key string, val []byte, ttl time.Duration) error {
	_, err := n.kapi.Set(n.context, key, string(val), &etcd.SetOptions{TTL: ttl})
	return convertErr(err)
}

func (n *ng) getJSONVal(key string, in interface{}) error {
	val, err := n.getVal(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), in)
}

func (n *ng) getVal(key string) (string, error) {
	response, err := n.kapi.Get(n.context, key, &etcd.GetOptions{Recursive: false, Sort: false, Quorum: n.requireQuorum})
	if err != nil {
		return "", convertErr(err)
	}

	if isDir(response.Node) {
		return "", &engine.NotFoundError{Message: fmt.Sprintf("missing key: %s", key)}
	}
	return response.Node.Value, nil
}

func (n *ng) getDirs(keys ...string) ([]string, error) {
	var out []string
	response, err := n.kapi.Get(n.context, strings.Join(keys, "/"), &etcd.GetOptions{Recursive: true, Sort: true, Quorum: n.requireQuorum})
	if err != nil {
		if notFound(err) {
			return out, nil
		}
		return nil, err
	}

	if response == nil || !isDir(response.Node) {
		return out, nil
	}

	for _, srvNode := range response.Node.Nodes {
		if isDir(srvNode) {
			out = append(out, srvNode.Key)
		}
	}
	return out, nil
}

func (n *ng) getVals(keys ...string) ([]Pair, error) {
	var out []Pair
	response, err := n.kapi.Get(n.context, strings.Join(keys, "/"), &etcd.GetOptions{Recursive: true, Sort: true, Quorum: n.requireQuorum})
	if err != nil {
		if notFound(err) {
			return out, nil
		}
		return nil, err
	}

	if !isDir(response.Node) {
		return out, nil
	}

	for _, srvNode := range response.Node.Nodes {
		if !isDir(srvNode) {
			out = append(out, Pair{srvNode.Key, srvNode.Value})
		}
	}
	return out, nil
}

func (n *ng) checkKeyExists(key string) error {
	_, err := n.kapi.Get(n.context, key, &etcd.GetOptions{Recursive: false, Sort: false, Quorum: n.requireQuorum})
	return convertErr(err)
}

func (n *ng) deleteKey(key string) error {
	_, err := n.kapi.Delete(n.context, key, &etcd.DeleteOptions{Recursive: true})
	return convertErr(err)
}

type Pair struct {
	Key string
	Val string
}

func suffix(key string) string {
	vals := strings.Split(key, "/")
	return vals[len(vals)-1]
}

func join(keys ...string) string {
	return strings.Join(keys, "/")
}

func notFound(e error) bool {
	err, ok := e.(etcd.Error)
	return ok && err.Code == etcd.ErrorCodeKeyNotFound
}

func convertErr(e error) error {
	if e == nil {
		return nil
	}
	switch err := e.(type) {
	case etcd.Error:
		if err.Code == etcd.ErrorCodeKeyNotFound {
			return &engine.NotFoundError{Message: err.Error()}
		}
		if err.Code == etcd.ErrorCodeNodeExist {
			return &engine.AlreadyExistsError{Message: err.Error()}
		}
	}
	return e
}

func isDir(n *etcd.Node) bool {
	return n != nil && n.Dir == true
}

const encryptionSecretBox = "secretbox.v1"

func responseToString(r *etcd.Response) string {
	return fmt.Sprintf("%s %s %d", r.Action, r.Node.Key, r.Index)
}

const (
	createA = "create"
	setA    = "set"
	deleteA = "delete"
	expireA = "expire"
	updateA = "update"
	cswapA  = "compareAndSwap"
	noTTL   = 0
)

type host struct {
	Name     string
	Settings hostSettings
}

type hostSettings struct {
	Default bool
	KeyPair []byte
	OCSP    engine.OCSPSettings
}
