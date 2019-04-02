package tls

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/sirupsen/logrus"
)

const (
	certificateHeader = "-----BEGIN CERTIFICATE-----\n"
)

// ClientCA defines traefik CA files for a entryPoint
// and it indicates if they are mandatory or have just to be analyzed if provided
type ClientCA struct {
	Files    FilesOrContents
	Optional bool
}

// TLS configures TLS for an entry point
type TLS struct {
	MinVersion         string `export:"true"`
	CipherSuites       []string
	Certificates       Certificates
	ClientCAFiles      FilesOrContents // Deprecated
	ClientCA           ClientCA
	DefaultCertificate *Certificate
	SniStrict          bool `export:"true"`
}

// FilesOrContents hold the CA we want to have in root
type FilesOrContents []FileOrContent

// Configuration allows mapping a TLS certificate to a list of entrypoints
type Configuration struct {
	EntryPoints []string
	Certificate *Certificate
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (r *FilesOrContents) String() string {
	sliceOfString := make([]string, len([]FileOrContent(*r)))
	for key, value := range *r {
		sliceOfString[key] = value.String()
	}
	return strings.Join(sliceOfString, ",")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (r *FilesOrContents) Set(value string) error {
	filesOrContents := strings.Split(value, ",")
	if len(filesOrContents) == 0 {
		return fmt.Errorf("bad FilesOrContents format: %s", value)
	}
	for _, fileOrContent := range filesOrContents {
		*r = append(*r, FileOrContent(fileOrContent))
	}
	return nil
}

// Get return the FilesOrContents list
func (r *FilesOrContents) Get() interface{} {
	return *r
}

// SetValue sets the FilesOrContents with val
func (r *FilesOrContents) SetValue(val interface{}) {
	*r = val.(FilesOrContents)
}

// Type is type of the struct
func (r *FilesOrContents) Type() string {
	return "filesorcontents"
}

// SortTLSPerEntryPoints converts TLS configuration sorted by Certificates into TLS configuration sorted by EntryPoints
func SortTLSPerEntryPoints(configurations []*Configuration, epConfiguration map[string]map[string]*tls.Certificate, defaultEntryPoints []string) {
	if epConfiguration == nil {
		epConfiguration = make(map[string]map[string]*tls.Certificate)
	}
	for _, conf := range configurations {
		if conf.EntryPoints == nil || len(conf.EntryPoints) == 0 {
			if log.GetLevel() >= logrus.DebugLevel {
				log.Debugf("No entryPoint is defined to add the certificate %s, it will be added to the default entryPoints: %s",
					conf.Certificate.getTruncatedCertificateName(),
					strings.Join(defaultEntryPoints, ", "))
			}
			conf.EntryPoints = append(conf.EntryPoints, defaultEntryPoints...)
		}
		for _, ep := range conf.EntryPoints {
			if err := conf.Certificate.AppendCertificate(epConfiguration, ep); err != nil {
				log.Errorf("Unable to append certificate %s to entrypoint %s: %v", conf.Certificate.getTruncatedCertificateName(), ep, err)
			}
		}
	}
}
