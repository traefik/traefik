package tls

import (
	"fmt"
	"strings"
)

const certificateHeader = "-----BEGIN CERTIFICATE-----\n"

// ClientCA defines traefik CA files for a entryPoint
// and it indicates if they are mandatory or have just to be analyzed if provided
type ClientCA struct {
	Files    FilesOrContents
	Optional bool
}

// TLS configures TLS for an entry point
type TLS struct {
	MinVersion   string `export:"true"`
	CipherSuites []string
	ClientCA     ClientCA
	SniStrict    bool `export:"true"`
}

// Store holds the options for a given Store
type Store struct {
	DefaultCertificate *Certificate
}

// FilesOrContents hold the CA we want to have in root
type FilesOrContents []FileOrContent

// Configuration allows mapping a TLS certificate to a list of entrypoints
type Configuration struct {
	Stores      []string
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
