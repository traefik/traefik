package provider

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/autogen"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// Provider defines methods of a provider.
type Provider interface {
	// Provide allows the provider to provide configurations to traefik
	// using the given configuration channel.
	Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error
}

// BaseProvider should be inherited by providers
type BaseProvider struct {
	Watch       bool              `description:"Watch provider"`
	Filename    string            `description:"Override default configuration template. For advanced users :)"`
	Constraints types.Constraints `description:"Filter services by constraint, matching with Traefik tags."`
}

// MatchConstraints must match with EVERY single contraint
// returns first constraint that do not match or nil
func (p *BaseProvider) MatchConstraints(tags []string) (bool, *types.Constraint) {
	// if there is no tags and no contraints, filtering is disabled
	if len(tags) == 0 && len(p.Constraints) == 0 {
		return true, nil
	}

	for _, constraint := range p.Constraints {
		// xor: if ok and constraint.MustMatch are equal, then no tag is currently matching with the constraint
		if ok := constraint.MatchConstraintWithAtLeastOneTag(tags); ok != constraint.MustMatch {
			return false, constraint
		}
	}

	// If no constraint or every constraints matching
	return true, nil
}

// GetConfiguration return the provider configuration using templating
func (p *BaseProvider) GetConfiguration(defaultTemplateFile string, funcMap template.FuncMap, templateObjects interface{}) (*types.Configuration, error) {
	var (
		buf []byte
		err error
	)
	configuration := new(types.Configuration)
	var defaultFuncMap = template.FuncMap{
		"replace":   Replace,
		"tolower":   strings.ToLower,
		"normalize": Normalize,
		"split":     split,
		"contains":  contains,
	}

	for funcID, funcElement := range funcMap {
		defaultFuncMap[funcID] = funcElement
	}

	tmpl := template.New(p.Filename).Funcs(defaultFuncMap)
	if len(p.Filename) > 0 {
		buf, err = ioutil.ReadFile(p.Filename)
		if err != nil {
			return nil, err
		}
	} else {
		buf, err = autogen.Asset(defaultTemplateFile)
		if err != nil {
			return nil, err
		}
	}
	_, err = tmpl.Parse(string(buf))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		return nil, err
	}

	var renderedTemplate = buffer.String()
	//	log.Debugf("Rendering results of %s:\n%s", defaultTemplateFile, renderedTemplate)
	if _, err := toml.Decode(renderedTemplate, configuration); err != nil {
		return nil, err
	}
	return configuration, nil
}

// Replace is an alias for strings.Replace
func Replace(s1 string, s2 string, s3 string) string {
	return strings.Replace(s3, s1, s2, -1)
}

func contains(substr, s string) bool {
	return strings.Contains(s, substr)
}

func split(sep, s string) []string {
	return strings.Split(s, sep)
}

// Normalize transform a string that work with the rest of traefik
func Normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}

// ReverseStringSlice invert the order of the given slice of string
func ReverseStringSlice(slice *[]string) {
	for i, j := 0, len(*slice)-1; i < j; i, j = i+1, j-1 {
		(*slice)[i], (*slice)[j] = (*slice)[j], (*slice)[i]
	}
}

// ClientTLS holds TLS specific configurations as client
// CA, Cert and Key can be either path or file contents
type ClientTLS struct {
	CA                 string `description:"TLS CA"`
	Cert               string `description:"TLS cert"`
	Key                string `description:"TLS key"`
	InsecureSkipVerify bool   `description:"TLS insecure skip verify"`
}

// CreateTLSConfig creates a TLS config from ClientTLS structures
func (clientTLS *ClientTLS) CreateTLSConfig() (*tls.Config, error) {
	var err error
	if clientTLS == nil {
		log.Warnf("clientTLS is nil")
		return nil, nil
	}
	caPool := x509.NewCertPool()
	if clientTLS.CA != "" {
		var ca []byte
		if _, errCA := os.Stat(clientTLS.CA); errCA == nil {
			ca, err = ioutil.ReadFile(clientTLS.CA)
			if err != nil {
				return nil, fmt.Errorf("Failed to read CA. %s", err)
			}
		} else {
			ca = []byte(clientTLS.CA)
		}
		caPool.AppendCertsFromPEM(ca)
	}

	cert := tls.Certificate{}
	_, errKeyIsFile := os.Stat(clientTLS.Key)

	if _, errCertIsFile := os.Stat(clientTLS.Cert); errCertIsFile == nil {
		if errKeyIsFile == nil {
			cert, err = tls.LoadX509KeyPair(clientTLS.Cert, clientTLS.Key)
			if err != nil {
				return nil, fmt.Errorf("Failed to load TLS keypair: %v", err)
			}
		} else {
			return nil, fmt.Errorf("tls cert is a file, but tls key is not")
		}
	} else {
		if errKeyIsFile != nil {
			cert, err = tls.X509KeyPair([]byte(clientTLS.Cert), []byte(clientTLS.Key))
			if err != nil {
				return nil, fmt.Errorf("Failed to load TLS keypair: %v", err)

			}
		} else {
			return nil, fmt.Errorf("tls key is a file, but tls cert is not")
		}
	}

	TLSConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caPool,
		InsecureSkipVerify: clientTLS.InsecureSkipVerify,
	}
	return TLSConfig, nil
}
