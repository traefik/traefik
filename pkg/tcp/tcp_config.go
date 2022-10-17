package tcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"net"
	"reflect"
	"sync"
	// "time"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	// "golang.org/x/net/http2"

	"github.com/traefik/traefik/v2/pkg/log"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
)

// NewTcpManager creates a new RoundTripperManager.
func NewTcpManager() *TcpManager {
	return &TcpManager{
		Transport: make(map[string]*net.Conn),
		configs:   make(map[string]*dynamic.TCPServersTransport),
	}
}

// TcpManager handles roundtripper for the reverse proxy.
type TcpManager struct {
	rtLock    sync.RWMutex
	configs   map[string]*dynamic.TCPServersTransport
	Transport map[string]*net.Conn
}

// Update updates the roundtrippers configurations.
func (r *TcpManager) Update(address string, newConfigs map[string]*dynamic.TCPServersTransport) {
	r.rtLock.Lock()
	defer r.rtLock.Unlock()
	for configName, config := range r.configs {
		newConfig, ok := newConfigs[configName]
		if !ok {
			delete(r.configs, configName)
			delete(r.Transport, configName)

			continue
		}
		// manager := service.NewRoundTripperManager()
		if reflect.DeepEqual(newConfig, config) {
			continue
		}
		var err error
		r.Transport[configName], err = createTcptransport(address, newConfig)
		if err != nil {
			log.WithoutContext().Errorf("Could not configure TCP %s, fallback on default transport: %v", configName, err)
			*r.Transport[configName], err = net.Dial("tcp", address)
			if err != nil {
				return
			}
		}

	}

	for newConfigName := range newConfigs {
		if _, ok := r.configs[newConfigName]; ok {
			continue

		}
	}
	for newConfigName, newConfig := range newConfigs {
		if _, ok := r.configs[newConfigName]; ok {
			continue
		}

		var err error
		r.Transport[newConfigName], err = createTcptransport(address, newConfig)
		if err != nil {
			log.WithoutContext().Errorf("Could not configure TCP Transport %s, fallback on default transport: %v", newConfigName, err)
			*r.Transport[newConfigName], err = net.Dial("tcp", address)
			if err != nil {
				return
			}
		}
	}

	r.configs = newConfigs
}

// createTcptransport creates an initial tcp configurations configured with the Transport configuration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost in Traefik at this point in time.
// Setting this value to the default of 100 could lead to confusing behavior and backwards compatibility issues.
func createTcptransport(address string, cfg *dynamic.TCPServersTransport) (*net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return &conn, errors.New("no transport configuration given")
	}
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	// dialer := &net.Dialer{
	// 	Timeout:   30 * time.Second,
	// 	KeepAlive: 30 * time.Second,
	// }
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer listen.Close()

	for {
		conntcp, err := listen.AcceptTCP()
		if err != nil {
			return nil, err
		}
		err = conntcp.SetKeepAlive(cfg.KeepAlive)
		if err != nil {
			return nil, err
		}
		conn = tls.Server(conntcp, &tls.Config{
			ServerName:         cfg.ServerAddress,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			RootCAs:            createRootCACertPool(cfg.RootCAs),
			Certificates:       cfg.Certificates.GetCertificates(),
		})
		return &conn, nil
	}
}

// Get get a roundtripper by name.
func (t *TcpManager) Get(name string) (*net.Conn, error) {
	if len(name) == 0 {
		name = "default@internal"
	}

	t.rtLock.RLock()
	defer t.rtLock.RUnlock()

	if rt, ok := t.Transport[name]; ok {
		return rt, nil
	}
	return nil, fmt.Errorf("servers transport not found %s", name)
}

func createRootCACertPool(rootCAs []traefiktls.FileOrContent) *x509.CertPool {
	if len(rootCAs) == 0 {
		return nil
	}

	roots := x509.NewCertPool()

	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.WithoutContext().Error("Error while read RootCAs", err)
			continue
		}
		roots.AppendCertsFromPEM(certContent)
	}

	return roots
}
