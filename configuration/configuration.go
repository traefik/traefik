package configuration

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/provider/boltdb"
	"github.com/containous/traefik/provider/consul"
	"github.com/containous/traefik/provider/docker"
	"github.com/containous/traefik/provider/dynamodb"
	"github.com/containous/traefik/provider/ecs"
	"github.com/containous/traefik/provider/etcd"
	"github.com/containous/traefik/provider/eureka"
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/provider/kubernetes"
	"github.com/containous/traefik/provider/marathon"
	"github.com/containous/traefik/provider/mesos"
	"github.com/containous/traefik/provider/rancher"
	"github.com/containous/traefik/provider/web"
	"github.com/containous/traefik/provider/zk"
	"github.com/containous/traefik/types"
)

const (
	// DefaultHealthCheckInterval is the default health check interval.
	DefaultHealthCheckInterval = 30 * time.Second

	// DefaultDialTimeout when connecting to a backend server.
	DefaultDialTimeout = 30 * time.Second

	// DefaultIdleTimeout before closing an idle connection.
	DefaultIdleTimeout = 180 * time.Second
)

// GlobalConfiguration holds global configuration (with providers, etc.).
// It's populated from the traefik configuration file passed as an argument to the binary.
type GlobalConfiguration struct {
	GraceTimeOut              flaeg.Duration          `short:"g" description:"Duration to give active requests a chance to finish before Traefik stops"`
	Debug                     bool                    `short:"d" description:"Enable debug mode"`
	CheckNewVersion           bool                    `description:"Periodically check if a new version has been released"`
	AccessLogsFile            string                  `description:"(Deprecated) Access logs file"` // Deprecated
	AccessLog                 *types.AccessLog        `description:"Access log settings"`
	TraefikLogsFile           string                  `description:"Traefik logs file. Stdout is used when omitted or empty"`
	LogLevel                  string                  `short:"l" description:"Log level"`
	EntryPoints               EntryPoints             `description:"Entrypoints definition using format: --entryPoints='Name:http Address::8000 Redirect.EntryPoint:https' --entryPoints='Name:https Address::4442 TLS:tests/traefik.crt,tests/traefik.key;prod/traefik.crt,prod/traefik.key'"`
	Cluster                   *types.Cluster          `description:"Enable clustering"`
	Constraints               types.Constraints       `description:"Filter services by constraint, matching with service tags"`
	ACME                      *acme.ACME              `description:"Enable ACME (Let's Encrypt): automatic SSL"`
	DefaultEntryPoints        DefaultEntryPoints      `description:"Entrypoints to be used by frontends that do not specify any entrypoint"`
	ProvidersThrottleDuration flaeg.Duration          `description:"Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time."`
	MaxIdleConnsPerHost       int                     `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host.  If zero, DefaultMaxIdleConnsPerHost is used"`
	IdleTimeout               flaeg.Duration          `description:"(Deprecated) maximum amount of time an idle (keep-alive) connection will remain idle before closing itself."` // Deprecated
	InsecureSkipVerify        bool                    `description:"Disable SSL certificate verification"`
	RootCAs                   RootCAs                 `description:"Add cert file for self-signed certicate"`
	Retry                     *Retry                  `description:"Enable retry sending request if network error"`
	HealthCheck               *HealthCheckConfig      `description:"Health check parameters"`
	RespondingTimeouts        *RespondingTimeouts     `description:"Timeouts for incoming requests to the Traefik instance"`
	ForwardingTimeouts        *ForwardingTimeouts     `description:"Timeouts for requests forwarded to the backend servers"`
	Docker                    *docker.Provider        `description:"Enable Docker backend with default settings"`
	File                      *file.Provider          `description:"Enable File backend with default settings"`
	Web                       *web.Provider           `description:"Enable Web backend with default settings"`
	Marathon                  *marathon.Provider      `description:"Enable Marathon backend with default settings"`
	Consul                    *consul.Provider        `description:"Enable Consul backend with default settings"`
	ConsulCatalog             *consul.CatalogProvider `description:"Enable Consul catalog backend with default settings"`
	Etcd                      *etcd.Provider          `description:"Enable Etcd backend with default settings"`
	Zookeeper                 *zk.Provider            `description:"Enable Zookeeper backend with default settings"`
	Boltdb                    *boltdb.Provider        `description:"Enable Boltdb backend with default settings"`
	Kubernetes                *kubernetes.Provider    `description:"Enable Kubernetes backend with default settings"`
	Mesos                     *mesos.Provider         `description:"Enable Mesos backend with default settings"`
	Eureka                    *eureka.Provider        `description:"Enable Eureka backend with default settings"`
	ECS                       *ecs.Provider           `description:"Enable ECS backend with default settings"`
	Rancher                   *rancher.Provider       `description:"Enable Rancher backend with default settings"`
	DynamoDB                  *dynamodb.Provider      `description:"Enable DynamoDB backend with default settings"`
}

// DefaultEntryPoints holds default entry points
type DefaultEntryPoints []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (dep *DefaultEntryPoints) String() string {
	return strings.Join(*dep, ",")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (dep *DefaultEntryPoints) Set(value string) error {
	entrypoints := strings.Split(value, ",")
	if len(entrypoints) == 0 {
		return fmt.Errorf("bad DefaultEntryPoints format: %s", value)
	}
	for _, entrypoint := range entrypoints {
		*dep = append(*dep, entrypoint)
	}
	return nil
}

// Get return the EntryPoints map
func (dep *DefaultEntryPoints) Get() interface{} {
	return DefaultEntryPoints(*dep)
}

// SetValue sets the EntryPoints map with val
func (dep *DefaultEntryPoints) SetValue(val interface{}) {
	*dep = DefaultEntryPoints(val.(DefaultEntryPoints))
}

// Type is type of the struct
func (dep *DefaultEntryPoints) Type() string {
	return "defaultentrypoints"
}

// RootCAs hold the CA we want to have in root
type RootCAs []FileOrContent

// FileOrContent hold a file path or content
type FileOrContent string

func (f FileOrContent) String() string {
	return string(f)
}

func (f FileOrContent) Read() ([]byte, error) {
	var content []byte
	if _, err := os.Stat(f.String()); err == nil {
		content, err = ioutil.ReadFile(f.String())
		if err != nil {
			return nil, err
		}
	} else {
		content = []byte(f)
	}
	return content, nil
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (r *RootCAs) String() string {
	sliceOfString := make([]string, len([]FileOrContent(*r)))
	for key, value := range *r {
		sliceOfString[key] = value.String()
	}
	return strings.Join(sliceOfString, ",")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (r *RootCAs) Set(value string) error {
	rootCAs := strings.Split(value, ",")
	if len(rootCAs) == 0 {
		return fmt.Errorf("bad RootCAs format: %s", value)
	}
	for _, rootCA := range rootCAs {
		*r = append(*r, FileOrContent(rootCA))
	}
	return nil
}

// Get return the EntryPoints map
func (r *RootCAs) Get() interface{} {
	return RootCAs(*r)
}

// SetValue sets the EntryPoints map with val
func (r *RootCAs) SetValue(val interface{}) {
	*r = RootCAs(val.(RootCAs))
}

// Type is type of the struct
func (r *RootCAs) Type() string {
	return "rootcas"
}

// EntryPoints holds entry points configuration of the reverse proxy (ip, port, TLS...)
type EntryPoints map[string]*EntryPoint

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (ep *EntryPoints) String() string {
	return fmt.Sprintf("%+v", *ep)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (ep *EntryPoints) Set(value string) error {
	regex := regexp.MustCompile(`(?:Name:(?P<Name>\S*))\s*(?:Address:(?P<Address>\S*))?\s*(?:TLS:(?P<TLS>\S*))?\s*((?P<TLSACME>TLS))?\s*(?:CA:(?P<CA>\S*))?\s*(?:Redirect.EntryPoint:(?P<RedirectEntryPoint>\S*))?\s*(?:Redirect.Regex:(?P<RedirectRegex>\\S*))?\s*(?:Redirect.Replacement:(?P<RedirectReplacement>\S*))?\s*(?:Compress:(?P<Compress>\S*))?\s*(?:WhiteListSourceRange:(?P<WhiteListSourceRange>\S*))?\s*(?:ProxyProtocol:(?P<ProxyProtocol>\S*))?`)
	match := regex.FindAllStringSubmatch(value, -1)
	if match == nil {
		return fmt.Errorf("bad EntryPoints format: %s", value)
	}
	matchResult := match[0]
	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 {
			result[name] = matchResult[i]
		}
	}
	var configTLS *TLS
	if len(result["TLS"]) > 0 {
		certs := Certificates{}
		if err := certs.Set(result["TLS"]); err != nil {
			return err
		}
		configTLS = &TLS{
			Certificates: certs,
		}
	} else if len(result["TLSACME"]) > 0 {
		configTLS = &TLS{
			Certificates: Certificates{},
		}
	}
	if len(result["CA"]) > 0 {
		files := strings.Split(result["CA"], ",")
		configTLS.ClientCAFiles = files
	}
	var redirect *Redirect
	if len(result["RedirectEntryPoint"]) > 0 || len(result["RedirectRegex"]) > 0 || len(result["RedirectReplacement"]) > 0 {
		redirect = &Redirect{
			EntryPoint:  result["RedirectEntryPoint"],
			Regex:       result["RedirectRegex"],
			Replacement: result["RedirectReplacement"],
		}
	}

	compress := false
	if len(result["Compress"]) > 0 {
		compress = strings.EqualFold(result["Compress"], "true") ||
			strings.EqualFold(result["Compress"], "enable") ||
			strings.EqualFold(result["Compress"], "on")
	}

	whiteListSourceRange := []string{}
	if len(result["WhiteListSourceRange"]) > 0 {
		whiteListSourceRange = strings.Split(result["WhiteListSourceRange"], ",")
	}

	proxyprotocol := false
	if len(result["ProxyProtocol"]) > 0 {
		proxyprotocol = strings.EqualFold(result["ProxyProtocol"], "true") ||
			strings.EqualFold(result["ProxyProtocol"], "enable") ||
			strings.EqualFold(result["ProxyProtocol"], "on")
	}

	(*ep)[result["Name"]] = &EntryPoint{
		Address:              result["Address"],
		TLS:                  configTLS,
		Redirect:             redirect,
		Compress:             compress,
		WhitelistSourceRange: whiteListSourceRange,
		ProxyProtocol:        proxyprotocol,
	}

	return nil
}

// Get return the EntryPoints map
func (ep *EntryPoints) Get() interface{} {
	return EntryPoints(*ep)
}

// SetValue sets the EntryPoints map with val
func (ep *EntryPoints) SetValue(val interface{}) {
	*ep = EntryPoints(val.(EntryPoints))
}

// Type is type of the struct
func (ep *EntryPoints) Type() string {
	return "entrypoints"
}

// EntryPoint holds an entry point configuration of the reverse proxy (ip, port, TLS...)
type EntryPoint struct {
	Network              string
	Address              string
	TLS                  *TLS
	Redirect             *Redirect
	Auth                 *types.Auth
	WhitelistSourceRange []string
	Compress             bool
	ProxyProtocol        bool
}

// Redirect configures a redirection of an entry point to another, or to an URL
type Redirect struct {
	EntryPoint  string
	Regex       string
	Replacement string
}

// TLS configures TLS for an entry point
type TLS struct {
	MinVersion    string
	CipherSuites  []string
	Certificates  Certificates
	ClientCAFiles []string
}

// MinVersion Map of allowed TLS minimum versions
var MinVersion = map[string]uint16{
	`VersionTLS10`: tls.VersionTLS10,
	`VersionTLS11`: tls.VersionTLS11,
	`VersionTLS12`: tls.VersionTLS12,
}

// CipherSuites Map of TLS CipherSuites from crypto/tls
// Available CipherSuites defined at https://golang.org/pkg/crypto/tls/#pkg-constants
var CipherSuites = map[string]uint16{
	`TLS_RSA_WITH_RC4_128_SHA`:                tls.TLS_RSA_WITH_RC4_128_SHA,
	`TLS_RSA_WITH_3DES_EDE_CBC_SHA`:           tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	`TLS_RSA_WITH_AES_128_CBC_SHA`:            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	`TLS_RSA_WITH_AES_256_CBC_SHA`:            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	`TLS_RSA_WITH_AES_128_CBC_SHA256`:         tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	`TLS_RSA_WITH_AES_128_GCM_SHA256`:         tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	`TLS_RSA_WITH_AES_256_GCM_SHA384`:         tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	`TLS_ECDHE_ECDSA_WITH_RC4_128_SHA`:        tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	`TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA`:    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	`TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA`:    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	`TLS_ECDHE_RSA_WITH_RC4_128_SHA`:          tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	`TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA`:     tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	`TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA`:      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	`TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA`:      tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	`TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256`: tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	`TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256`:   tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	`TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`:   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	`TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256`: tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	`TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`:   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	`TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384`: tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	`TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305`:    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	`TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305`:  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
}

// Certificates defines traefik certificates type
// Certs and Keys could be either a file path, or the file content itself
type Certificates []Certificate

//CreateTLSConfig creates a TLS config from Certificate structures
func (certs *Certificates) CreateTLSConfig() (*tls.Config, error) {
	config := &tls.Config{}
	config.Certificates = []tls.Certificate{}
	certsSlice := []Certificate(*certs)
	for _, v := range certsSlice {
		cert := tls.Certificate{}

		var err error

		certContent, err := v.CertFile.Read()
		if err != nil {
			return nil, err
		}

		keyContent, err := v.KeyFile.Read()
		if err != nil {
			return nil, err
		}

		cert, err = tls.X509KeyPair(certContent, keyContent)
		if err != nil {
			return nil, err
		}

		config.Certificates = append(config.Certificates, cert)
	}
	return config, nil
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (certs *Certificates) String() string {
	if len(*certs) == 0 {
		return ""
	}
	var result []string
	for _, certificate := range *certs {
		result = append(result, certificate.CertFile.String()+","+certificate.KeyFile.String())
	}
	return strings.Join(result, ";")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (certs *Certificates) Set(value string) error {
	certificates := strings.Split(value, ";")
	for _, certificate := range certificates {
		files := strings.Split(certificate, ",")
		if len(files) != 2 {
			return fmt.Errorf("bad certificates format: %s", value)
		}
		*certs = append(*certs, Certificate{
			CertFile: FileOrContent(files[0]),
			KeyFile:  FileOrContent(files[1]),
		})
	}
	return nil
}

// Type is type of the struct
func (certs *Certificates) Type() string {
	return "certificates"
}

// Certificate holds a SSL cert/key pair
// Certs and Key could be either a file path, or the file content itself
type Certificate struct {
	CertFile FileOrContent
	KeyFile  FileOrContent
}

// Retry contains request retry config
type Retry struct {
	Attempts int `description:"Number of attempts"`
}

// HealthCheckConfig contains health check configuration parameters.
type HealthCheckConfig struct {
	Interval flaeg.Duration `description:"Default periodicity of enabled health checks"`
}

// RespondingTimeouts contains timeout configurations for incoming requests to the Traefik instance.
type RespondingTimeouts struct {
	ReadTimeout  flaeg.Duration `description:"ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set"`
	WriteTimeout flaeg.Duration `description:"WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set"`
	IdleTimeout  flaeg.Duration `description:"IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. Defaults to 180 seconds. If zero, no timeout is set"`
}

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           flaeg.Duration `description:"The amount of time to wait until a connection to a backend server can be established. Defaults to 30 seconds. If zero, no timeout exists"`
	ResponseHeaderTimeout flaeg.Duration `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists"`
}
