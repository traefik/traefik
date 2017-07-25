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

package integration

import (
	"io"
	"os"
	"path/filepath"

	"github.com/coreos/etcd/pkg/transport"
)

// copyTLSFiles clones certs files to dst directory.
func copyTLSFiles(ti transport.TLSInfo, dst string) (transport.TLSInfo, error) {
	ci := transport.TLSInfo{
		KeyFile:        filepath.Join(dst, "server-key.pem"),
		CertFile:       filepath.Join(dst, "server.pem"),
		TrustedCAFile:  filepath.Join(dst, "etcd-root-ca.pem"),
		ClientCertAuth: ti.ClientCertAuth,
	}
	if err := copyFile(ti.KeyFile, ci.KeyFile); err != nil {
		return transport.TLSInfo{}, err
	}
	if err := copyFile(ti.CertFile, ci.CertFile); err != nil {
		return transport.TLSInfo{}, err
	}
	if err := copyFile(ti.TrustedCAFile, ci.TrustedCAFile); err != nil {
		return transport.TLSInfo{}, err
	}
	return ci, nil
}

func copyFile(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err = io.Copy(w, f); err != nil {
		return err
	}
	return w.Sync()
}
