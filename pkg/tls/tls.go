package tls

const certificateHeader = "-----BEGIN CERTIFICATE-----\n"

// ClientCA defines traefik CA files for a entryPoint
// and it indicates if they are mandatory or have just to be analyzed if provided.
type ClientCA struct {
	Files    []FileOrContent `json:"files,omitempty" toml:"files,omitempty" yaml:"files,omitempty"`
	Optional bool            `json:"optional,omitempty" toml:"optional,omitempty" yaml:"optional,omitempty"`
}

// Options configures TLS for an entry point
type Options struct {
	MinVersion   string   `json:"minVersion,omitempty" toml:"minVersion,omitempty" yaml:"minVersion,omitempty" export:"true"`
	CipherSuites []string `json:"cipherSuites,omitempty" toml:"cipherSuites,omitempty" yaml:"cipherSuites,omitempty"`
	ClientCA     ClientCA `json:"clientCA,omitempty" toml:"clientCA,omitempty" yaml:"clientCA,omitempty"`
	SniStrict    bool     `json:"sniStrict,omitempty" toml:"sniStrict,omitempty" yaml:"sniStrict,omitempty" export:"true"`
}

// Store holds the options for a given Store
type Store struct {
	DefaultCertificate *Certificate `json:"defaultCertificate,omitempty" toml:"defaultCertificate,omitempty" yaml:"defaultCertificate,omitempty"`
}

// CertAndStores allows mapping a TLS certificate to a list of entry points.
type CertAndStores struct {
	Certificate `yaml:",inline"`
	Stores      []string `json:"stores,omitempty" toml:"stores,omitempty" yaml:"stores,omitempty"`
}
