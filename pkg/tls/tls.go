package tls

const certificateHeader = "-----BEGIN CERTIFICATE-----\n"

// ClientCA defines traefik CA files for a entryPoint
// and it indicates if they are mandatory or have just to be analyzed if provided.
type ClientCA struct {
	Files    []FileOrContent
	Optional bool
}

// Options configures TLS for an entry point
type Options struct {
	MinVersion   string `export:"true"`
	CipherSuites []string
	ClientCA     ClientCA
	SniStrict    bool `export:"true"`
}

// Store holds the options for a given Store
type Store struct {
	DefaultCertificate *Certificate
}

// Configuration allows mapping a TLS certificate to a list of entry points.
// FIXME better name?
type Configuration struct {
	Certificate `yaml:",inline"`
	Stores      []string
}
