// Copyright 2017 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package yaml handles yaml-formatted clientv3 configuration data.
package yaml

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/ghodss/yaml"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/tlsutil"
)

type yamlConfig struct {
	clientv3.Config

	InsecureTransport     bool   `json:"insecure-transport"`
	InsecureSkipTLSVerify bool   `json:"insecure-skip-tls-verify"`
	Certfile              string `json:"cert-file"`
	Keyfile               string `json:"key-file"`
	CAfile                string `json:"ca-file"`
}

// NewConfig creates a new clientv3.Config from a yaml file.
func NewConfig(fpath string) (*clientv3.Config, error) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	yc := &yamlConfig{}

	err = yaml.Unmarshal(b, yc)
	if err != nil {
		return nil, err
	}

	if yc.InsecureTransport {
		return &yc.Config, nil
	}

	var (
		cert *tls.Certificate
		cp   *x509.CertPool
	)

	if yc.Certfile != "" && yc.Keyfile != "" {
		cert, err = tlsutil.NewCert(yc.Certfile, yc.Keyfile, nil)
		if err != nil {
			return nil, err
		}
	}

	if yc.CAfile != "" {
		cp, err = tlsutil.NewCertPool([]string{yc.CAfile})
		if err != nil {
			return nil, err
		}
	}

	tlscfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: yc.InsecureSkipTLSVerify,
		RootCAs:            cp,
	}
	if cert != nil {
		tlscfg.Certificates = []tls.Certificate{*cert}
	}
	yc.Config.TLS = tlscfg

	return &yc.Config, nil
}
