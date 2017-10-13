package configuration

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
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
	GraceTimeOut              flaeg.Duration          `short:"g" description:"Duration to give active requests a chance to finish before Traefik stops" export:"true"`
	Debug                     bool                    `short:"d" description:"Enable debug mode" export:"true"`
	CheckNewVersion           bool                    `description:"Periodically check if a new version has been released" export:"true"`
	AccessLogsFile            string                  `description:"(Deprecated) Access logs file" export:"true"` // Deprecated
	AccessLog                 *types.AccessLog        `description:"Access log settings" export:"true"`
	TraefikLogsFile           string                  `description:"Traefik logs file. Stdout is used when omitted or empty" export:"true"`
	LogLevel                  string                  `short:"l" description:"Log level" export:"true"`
	EntryPoints               EntryPoints             `description:"Entrypoints definition using format: --entryPoints='Name:http Address::8000 Redirect.EntryPoint:https' --entryPoints='Name:https Address::4442 TLS:tests/traefik.crt,tests/traefik.key;prod/traefik.crt,prod/traefik.key'" export:"true"`
	Cluster                   *types.Cluster          `description:"Enable clustering" export:"true"`
	Constraints               types.Constraints       `description:"Filter services by constraint, matching with service tags" export:"true"`
	ACME                      *acme.ACME              `description:"Enable ACME (Let's Encrypt): automatic SSL" export:"true"`
	DefaultEntryPoints        DefaultEntryPoints      `description:"Entrypoints to be used by frontends that do not specify any entrypoint" export:"true"`
	ProvidersThrottleDuration flaeg.Duration          `description:"Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time." export:"true"`
	MaxIdleConnsPerHost       int                     `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host.  If zero, DefaultMaxIdleConnsPerHost is used" export:"true"`
	IdleTimeout               flaeg.Duration          `description:"(Deprecated) maximum amount of time an idle (keep-alive) connection will remain idle before closing itself." export:"true"` // Deprecated
	InsecureSkipVerify        bool                    `description:"Disable SSL certificate verification" export:"true"`
	RootCAs                   RootCAs                 `description:"Add cert file for self-signed certificate"`
	Retry                     *Retry                  `description:"Enable retry sending request if network error" export:"true"`
	HealthCheck               *HealthCheckConfig      `description:"Health check parameters" export:"true"`
	RespondingTimeouts        *RespondingTimeouts     `description:"Timeouts for incoming requests to the Traefik instance" export:"true"`
	ForwardingTimeouts        *ForwardingTimeouts     `description:"Timeouts for requests forwarded to the backend servers" export:"true"`
	Docker                    *docker.Provider        `description:"Enable Docker backend with default settings" export:"true"`
	File                      *file.Provider          `description:"Enable File backend with default settings" export:"true"`
	Web                       *web.Provider           `description:"Enable Web backend with default settings" export:"true"`
	Marathon                  *marathon.Provider      `description:"Enable Marathon backend with default settings" export:"true"`
	Consul                    *consul.Provider        `description:"Enable Consul backend with default settings" export:"true"`
	ConsulCatalog             *consul.CatalogProvider `description:"Enable Consul catalog backend with default settings" export:"true"`
	Etcd                      *etcd.Provider          `description:"Enable Etcd backend with default settings" export:"true"`
	Zookeeper                 *zk.Provider            `description:"Enable Zookeeper backend with default settings" export:"true"`
	Boltdb                    *boltdb.Provider        `description:"Enable Boltdb backend with default settings" export:"true"`
	Kubernetes                *kubernetes.Provider    `description:"Enable Kubernetes backend with default settings" export:"true"`
	Mesos                     *mesos.Provider         `description:"Enable Mesos backend with default settings" export:"true"`
	Eureka                    *eureka.Provider        `description:"Enable Eureka backend with default settings" export:"true"`
	ECS                       *ecs.Provider           `description:"Enable ECS backend with default settings" export:"true"`
	Rancher                   *rancher.Provider       `description:"Enable Rancher backend with default settings" export:"true"`
	DynamoDB                  *dynamodb.Provider      `description:"Enable DynamoDB backend with default settings" export:"true"`
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
	result := parseEntryPointsConfiguration(value)

	var configTLS *TLS
	if len(result["tls"]) > 0 {
		certs := Certificates{}
		if err := certs.Set(result["tls"]); err != nil {
			return err
		}
		configTLS = &TLS{
			Certificates: certs,
		}
	} else if len(result["tls_acme"]) > 0 {
		configTLS = &TLS{
			Certificates: Certificates{},
		}
	}
	if len(result["ca"]) > 0 {
		files := strings.Split(result["ca"], ",")
		configTLS.ClientCAFiles = files
	}
	var redirect *Redirect
	if len(result["redirect_entrypoint"]) > 0 || len(result["redirect_regex"]) > 0 || len(result["redirect_replacement"]) > 0 {
		redirect = &Redirect{
			EntryPoint:  result["redirect_entrypoint"],
			Regex:       result["redirect_regex"],
			Replacement: result["redirect_replacement"],
		}
	}

	whiteListSourceRange := []string{}
	if len(result["whitelistsourcerange"]) > 0 {
		whiteListSourceRange = strings.Split(result["whitelistsourcerange"], ",")
	}

	compress := toBool(result, "compress")

	var proxyProtocol *ProxyProtocol
	if len(result["proxyprotocol_trustedips"]) > 0 {
		trustedIPs := strings.Split(result["proxyprotocol_trustedips"], ",")
		proxyProtocol = &ProxyProtocol{
			TrustedIPs: trustedIPs,
		}
	}

	(*ep)[result["name"]] = &EntryPoint{
		Address:              result["address"],
		TLS:                  configTLS,
		Redirect:             redirect,
		Compress:             compress,
		WhitelistSourceRange: whiteListSourceRange,
		ProxyProtocol:        proxyProtocol,
	}

	return nil
}

func parseEntryPointsConfiguration(raw string) map[string]string {
	sections := strings.Fields(raw)

	config := make(map[string]string)
	for _, part := range sections {
		field := strings.SplitN(part, ":", 2)
		name := strings.ToLower(strings.Replace(field[0], ".", "_", -1))
		if len(field) > 1 {
			config[name] = field[1]
		} else {
			if strings.EqualFold(name, "TLS") {
				config["tls_acme"] = "TLS"
			} else {
				config[name] = ""
			}
		}
	}
	return config
}

func toBool(conf map[string]string, key string) bool {
	if val, ok := conf[key]; ok {
		return strings.EqualFold(val, "true") ||
			strings.EqualFold(val, "enable") ||
			strings.EqualFold(val, "on")
	}
	return false
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
	TLS                  *TLS        `export:"true"`
	Redirect             *Redirect   `export:"true"`
	Auth                 *types.Auth `export:"true"`
	WhitelistSourceRange []string
	Compress             bool           `export:"true"`
	ProxyProtocol        *ProxyProtocol `export:"true"`
}

// Redirect configures a redirection of an entry point to another, or to an URL
type Redirect struct {
	EntryPoint  string
	Regex       string
	Replacement string
}

// TLS configures TLS for an entry point
type TLS struct {
	MinVersion    string `export:"true"`
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
		var err error

		certContent, err := v.CertFile.Read()
		if err != nil {
			return nil, err
		}

		keyContent, err := v.KeyFile.Read()
		if err != nil {
			return nil, err
		}

		cert, err := tls.X509KeyPair(certContent, keyContent)
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
	Attempts int `description:"Number of attempts" export:"true"`
}

// HealthCheckConfig contains health check configuration parameters.
type HealthCheckConfig struct {
	Interval flaeg.Duration `description:"Default periodicity of enabled health checks" export:"true"`
}

// RespondingTimeouts contains timeout configurations for incoming requests to the Traefik instance.
type RespondingTimeouts struct {
	ReadTimeout  flaeg.Duration `description:"ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set" export:"true"`
	WriteTimeout flaeg.Duration `description:"WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set" export:"true"`
	IdleTimeout  flaeg.Duration `description:"IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. Defaults to 180 seconds. If zero, no timeout is set" export:"true"`
}

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           flaeg.Duration `description:"The amount of time to wait until a connection to a backend server can be established. Defaults to 30 seconds. If zero, no timeout exists" export:"true"`
	ResponseHeaderTimeout flaeg.Duration `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists" export:"true"`
}

// ProxyProtocol contains Proxy-Protocol configuration
type ProxyProtocol struct {
	TrustedIPs []string
}
