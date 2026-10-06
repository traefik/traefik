package tls

import (
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/types"
)

const (
	//////////////////////
	//
	//   ClientAuthType
	//
	/////////////////////

	// NoClientCert indicates that no client certificate should be requested
	// during the handshake, and if any certificates are sent they will not
	// be verified.
	NoClientCert = "NoClientCert"
	// RequestClientCert indicates that a client certificate should be requested
	// during the handshake, but does not require that the client send any
	// certificates.
	RequestClientCert = "RequestClientCert"
	// RequireAnyClientCert indicates that a client certificate should be requested
	// during the handshake, and that at least one certificate is required to be
	// sent by the client, but that certificate is not required to be valid.
	RequireAnyClientCert = "RequireAnyClientCert"
	// VerifyClientCertIfGiven indicates that a client certificate should be requested
	// during the handshake, but does not require that the client sends a
	// certificate. If the client does send a certificate it is required to be
	// valid.
	VerifyClientCertIfGiven = "VerifyClientCertIfGiven"
	// RequireAndVerifyClientCert indicates that a client certificate should be requested
	// during the handshake, and that at least one valid certificate is required
	// to be sent by the client.
	RequireAndVerifyClientCert = "RequireAndVerifyClientCert"
	// RequireAndVerifyClientCert indicates that a client certificate should be requested
	// during the handshake, and that at least one valid certificate is required
	// to be sent by the client.
	RequireAndVerifyClientCertWithExpiry = "RequireAndVerifyClientCertWithExpiry"

	//////////////////////
	//
	//   CRL mode
	//
	//////////////////////

	// NOOP indicates that CRL validations will not be performed.
	CRLNOOP = "NOOP"

	// Lax indicates that CRL validation will only enforce CRL validation, for each certificates,
	// if a CRL attribute is present.
	CRLLax = "Lax"

	// Strict indicates that CRL validation will ensure CRL attributes are present
	// on client certs and check for each certificate in verified chain if they are expired
	// against their CRL file.
	CRLStrict = "Strict"

	//////////////////////
	//
	//   CRL mode
	//
	//////////////////////

	// Open indicates that when expired,
	// CRL loaded through HTTP should be refreshed by the first request using it,
	// other requests will be using stale data
	CRLExpirationOpen = "Open"

	// Open indicates that when expired,
	// CRL loaded through HTTP should be refreshed by the first request using it,
	// other requests will be locked (expect high latency burst)
	CRLExpirationFailedClosed = "FailedClosed"
)

// +k8s:deepcopy-gen=true

// CRLHTTPWhitelist defines how CDP whitelist should be handled for CRLs fetched through HTTP
type CRLHTTPWhitelist struct {
	// Enables whitelist handling, setting it to false may lead to high memory usage
	// due to a high number of CRL being stored
	Enabled bool `json:"enabled,omitempty" toml:"enabled,omitempty" yaml:"enabled,omitempty" export:"true"`
	// CRL DP whitelist
	DistributionPoints []string `json:"distributionPoints,omitempty" toml:"distributionPoints,omitempty" yaml:"distributionPoints,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// CRLHTTP defines the parameters of the client authentication regarding client certificate expiry based on CRLs
// when loading CRLs from HTTP endpoints.
type CRLHTTP struct {
	// Defines how CRL expiry is handled.
	// Available values are: "Open", "FailedClosed"
	ExpirationStrategy string `json:"expirationStrategy,omitempty" toml:"expirationStrategy,omitempty" yaml:"expirationStrategy,omitempty" export:"true"`
	// Timeout for HTTP requests used to fetch CRLs for this TLS Options.
	// If unset, a default timeout configured at the Traefik static level is used.
	Timeout ptypes.Duration `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	// Error backoff duration for HTTP requests when endpoints are in error
	// If unset, a default duration is set at TRaefik static configuration level
	ErrorBackoff ptypes.Duration `json:"errorBackoff,omitempty" toml:"errorBackoff,omitempty" yaml:"errorBackoff,omitempty" export:"true"`
	// Max CRL file size exprimed in bytes.
	// Default is 1MB
	MaxCRLBytes int64 `json:"maxCRLBytes,omitempty" toml:"maxCRLBytes,omitempty" yaml:"maxCRLBytes,omitempty" export:"true"`
	// CRL Distribution Point whitelist config
	Whitelist CRLHTTPWhitelist `json:"whitelist,omitempty" toml:"whitelist,omitempty" yaml:"whitelist,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// CRL defines the parameters of the client authentication regarding client certificate expiry based on CRLs.
type CRL struct {
	// Mode denotates how CRL validation will be handled.
	// Available values are: "NOOP", "Lax", "Strict".
	// Defaults to "NOOP".
	Mode string `json:"mode,omitempty" toml:"mode,omitempty" yaml:"mode,omitempty" export:"true"`
	// Interval between files reload, setting it to 0 will turn off reload
	ReloadInterval ptypes.Duration `json:"reloadInterval,omitempty" toml:"reloadInterval,omitempty" yaml:"reloadInterval,omitempty" export:"true"`
	// HTTP config for Load mode HTTP
	HTTP CRLHTTP `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Expiry defines the parameters of the client authentication regarding client certificate expiry.
type Expiry struct {
	// CRL based expiry configuration
	CRL CRL `json:"crl,omitempty" toml:"crl,omitempty" yaml:"crl,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ClientAuth defines the parameters of the client authentication part of the TLS connection, if any.
type ClientAuth struct {
	CAFiles []types.FileOrContent `json:"caFiles,omitempty" toml:"caFiles,omitempty" yaml:"caFiles,omitempty"`
	// ClientAuthType defines the client authentication type to apply.
	// The available values are: "NoClientCert", "RequestClientCert", "VerifyClientCertIfGiven" and "RequireAndVerifyClientCert".
	ClientAuthType string `json:"clientAuthType,omitempty" toml:"clientAuthType,omitempty" yaml:"clientAuthType,omitempty" export:"true"`
	// Client certificates expiry handling
	Expiry Expiry `json:"expiry,omitempty" toml:"expiry,omitempty" yaml:"expiry,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Options configures TLS for an entry point.
type Options struct {
	MinVersion            string     `json:"minVersion,omitempty" toml:"minVersion,omitempty" yaml:"minVersion,omitempty" export:"true"`
	MaxVersion            string     `json:"maxVersion,omitempty" toml:"maxVersion,omitempty" yaml:"maxVersion,omitempty" export:"true"`
	CipherSuites          []string   `json:"cipherSuites,omitempty" toml:"cipherSuites,omitempty" yaml:"cipherSuites,omitempty" export:"true"`
	CurvePreferences      []string   `json:"curvePreferences,omitempty" toml:"curvePreferences,omitempty" yaml:"curvePreferences,omitempty" export:"true"`
	ClientAuth            ClientAuth `json:"clientAuth,omitempty" toml:"clientAuth,omitempty" yaml:"clientAuth,omitempty" export:"true"`
	SniStrict             bool       `json:"sniStrict,omitempty" toml:"sniStrict,omitempty" yaml:"sniStrict,omitempty" export:"true"`
	ALPNProtocols         []string   `json:"alpnProtocols,omitempty" toml:"alpnProtocols,omitempty" yaml:"alpnProtocols,omitempty" export:"true"`
	DisableSessionTickets bool       `json:"disableSessionTickets,omitempty" toml:"disableSessionTickets,omitempty" yaml:"disableSessionTickets,omitempty" export:"true"`

	// Deprecated: https://github.com/golang/go/issues/45430
	PreferServerCipherSuites *bool `json:"preferServerCipherSuites,omitempty" toml:"preferServerCipherSuites,omitempty" yaml:"preferServerCipherSuites,omitempty" export:"true"`
}

// SetDefaults sets the default values for an Options struct.
func (o *Options) SetDefaults() {
	// ensure http2 enabled
	o.ALPNProtocols = DefaultTLSOptions.ALPNProtocols
	o.CipherSuites = DefaultTLSOptions.CipherSuites
}

// +k8s:deepcopy-gen=true

// Store holds the options for a given Store.
type Store struct {
	DefaultCertificate   *Certificate   `json:"defaultCertificate,omitempty" toml:"defaultCertificate,omitempty" yaml:"defaultCertificate,omitempty" label:"-" export:"true"`
	DefaultGeneratedCert *GeneratedCert `json:"defaultGeneratedCert,omitempty" toml:"defaultGeneratedCert,omitempty" yaml:"defaultGeneratedCert,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// GeneratedCert defines the default generated certificate configuration.
type GeneratedCert struct {
	// Resolver is the name of the resolver that will be used to issue the DefaultCertificate.
	Resolver string `json:"resolver,omitempty" toml:"resolver,omitempty" yaml:"resolver,omitempty" export:"true"`
	// Domain is the domain definition for the DefaultCertificate.
	Domain *types.Domain `json:"domain,omitempty" toml:"domain,omitempty" yaml:"domain,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// CertAndStores allows mapping a TLS certificate to a list of entry points.
type CertAndStores struct {
	Certificate `yaml:",inline" export:"true"`

	Stores []string `json:"stores,omitempty" toml:"stores,omitempty" yaml:"stores,omitempty" export:"true"`
}
