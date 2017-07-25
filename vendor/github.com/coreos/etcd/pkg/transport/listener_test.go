// Copyright 2015 The etcd Authors
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

package transport

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func createSelfCert() (*TLSInfo, func(), error) {
	d, terr := ioutil.TempDir("", "etcd-test-tls-")
	if terr != nil {
		return nil, nil, terr
	}
	info, err := SelfCert(d, []string{"127.0.0.1"})
	if err != nil {
		return nil, nil, err
	}
	return &info, func() { os.RemoveAll(d) }, nil
}

func fakeCertificateParserFunc(cert tls.Certificate, err error) func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
	return func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
		return cert, err
	}
}

// TestNewListenerTLSInfo tests that NewListener with valid TLSInfo returns
// a TLS listener that accepts TLS connections.
func TestNewListenerTLSInfo(t *testing.T) {
	tlsInfo, del, err := createSelfCert()
	if err != nil {
		t.Fatalf("unable to create cert: %v", err)
	}
	defer del()
	testNewListenerTLSInfoAccept(t, *tlsInfo)
}

func testNewListenerTLSInfoAccept(t *testing.T, tlsInfo TLSInfo) {
	ln, err := NewListener("127.0.0.1:0", "https", &tlsInfo)
	if err != nil {
		t.Fatalf("unexpected NewListener error: %v", err)
	}
	defer ln.Close()

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	cli := &http.Client{Transport: tr}
	go cli.Get("https://" + ln.Addr().String())

	conn, err := ln.Accept()
	if err != nil {
		t.Fatalf("unexpected Accept error: %v", err)
	}
	defer conn.Close()
	if _, ok := conn.(*tls.Conn); !ok {
		t.Errorf("failed to accept *tls.Conn")
	}
}

func TestNewListenerTLSEmptyInfo(t *testing.T) {
	_, err := NewListener("127.0.0.1:0", "https", nil)
	if err == nil {
		t.Errorf("err = nil, want not presented error")
	}
}

func TestNewTransportTLSInfo(t *testing.T) {
	tlsinfo, del, err := createSelfCert()
	if err != nil {
		t.Fatalf("unable to create cert: %v", err)
	}
	defer del()

	tests := []TLSInfo{
		{},
		{
			CertFile: tlsinfo.CertFile,
			KeyFile:  tlsinfo.KeyFile,
		},
		{
			CertFile: tlsinfo.CertFile,
			KeyFile:  tlsinfo.KeyFile,
			CAFile:   tlsinfo.CAFile,
		},
		{
			CAFile: tlsinfo.CAFile,
		},
	}

	for i, tt := range tests {
		tt.parseFunc = fakeCertificateParserFunc(tls.Certificate{}, nil)
		trans, err := NewTransport(tt, time.Second)
		if err != nil {
			t.Fatalf("Received unexpected error from NewTransport: %v", err)
		}

		if trans.TLSClientConfig == nil {
			t.Fatalf("#%d: want non-nil TLSClientConfig", i)
		}
	}
}

func TestTLSInfoNonexist(t *testing.T) {
	tlsInfo := TLSInfo{CertFile: "@badname", KeyFile: "@badname"}
	_, err := tlsInfo.ServerConfig()
	werr := &os.PathError{
		Op:   "open",
		Path: "@badname",
		Err:  errors.New("no such file or directory"),
	}
	if err.Error() != werr.Error() {
		t.Errorf("err = %v, want %v", err, werr)
	}
}

func TestTLSInfoEmpty(t *testing.T) {
	tests := []struct {
		info TLSInfo
		want bool
	}{
		{TLSInfo{}, true},
		{TLSInfo{CAFile: "baz"}, true},
		{TLSInfo{CertFile: "foo"}, false},
		{TLSInfo{KeyFile: "bar"}, false},
		{TLSInfo{CertFile: "foo", KeyFile: "bar"}, false},
		{TLSInfo{CertFile: "foo", CAFile: "baz"}, false},
		{TLSInfo{KeyFile: "bar", CAFile: "baz"}, false},
		{TLSInfo{CertFile: "foo", KeyFile: "bar", CAFile: "baz"}, false},
	}

	for i, tt := range tests {
		got := tt.info.Empty()
		if tt.want != got {
			t.Errorf("#%d: result of Empty() incorrect: want=%t got=%t", i, tt.want, got)
		}
	}
}

func TestTLSInfoMissingFields(t *testing.T) {
	tlsinfo, del, err := createSelfCert()
	if err != nil {
		t.Fatalf("unable to create cert: %v", err)
	}
	defer del()

	tests := []TLSInfo{
		{CertFile: tlsinfo.CertFile},
		{KeyFile: tlsinfo.KeyFile},
		{CertFile: tlsinfo.CertFile, CAFile: tlsinfo.CAFile},
		{KeyFile: tlsinfo.KeyFile, CAFile: tlsinfo.CAFile},
	}

	for i, info := range tests {
		if _, err = info.ServerConfig(); err == nil {
			t.Errorf("#%d: expected non-nil error from ServerConfig()", i)
		}

		if _, err = info.ClientConfig(); err == nil {
			t.Errorf("#%d: expected non-nil error from ClientConfig()", i)
		}
	}
}

func TestTLSInfoParseFuncError(t *testing.T) {
	tlsinfo, del, err := createSelfCert()
	if err != nil {
		t.Fatalf("unable to create cert: %v", err)
	}
	defer del()

	tlsinfo.parseFunc = fakeCertificateParserFunc(tls.Certificate{}, errors.New("fake"))

	if _, err = tlsinfo.ServerConfig(); err == nil {
		t.Errorf("expected non-nil error from ServerConfig()")
	}

	if _, err = tlsinfo.ClientConfig(); err == nil {
		t.Errorf("expected non-nil error from ClientConfig()")
	}
}

func TestTLSInfoConfigFuncs(t *testing.T) {
	tlsinfo, del, err := createSelfCert()
	if err != nil {
		t.Fatalf("unable to create cert: %v", err)
	}
	defer del()

	tests := []struct {
		info       TLSInfo
		clientAuth tls.ClientAuthType
		wantCAs    bool
	}{
		{
			info:       TLSInfo{CertFile: tlsinfo.CertFile, KeyFile: tlsinfo.KeyFile},
			clientAuth: tls.NoClientCert,
			wantCAs:    false,
		},

		{
			info:       TLSInfo{CertFile: tlsinfo.CertFile, KeyFile: tlsinfo.KeyFile, CAFile: tlsinfo.CertFile},
			clientAuth: tls.RequireAndVerifyClientCert,
			wantCAs:    true,
		},
	}

	for i, tt := range tests {
		tt.info.parseFunc = fakeCertificateParserFunc(tls.Certificate{}, nil)

		sCfg, err := tt.info.ServerConfig()
		if err != nil {
			t.Errorf("#%d: expected nil error from ServerConfig(), got non-nil: %v", i, err)
		}

		if tt.wantCAs != (sCfg.ClientCAs != nil) {
			t.Errorf("#%d: wantCAs=%t but ClientCAs=%v", i, tt.wantCAs, sCfg.ClientCAs)
		}

		cCfg, err := tt.info.ClientConfig()
		if err != nil {
			t.Errorf("#%d: expected nil error from ClientConfig(), got non-nil: %v", i, err)
		}

		if tt.wantCAs != (cCfg.RootCAs != nil) {
			t.Errorf("#%d: wantCAs=%t but RootCAs=%v", i, tt.wantCAs, sCfg.RootCAs)
		}
	}
}

func TestNewListenerUnixSocket(t *testing.T) {
	l, err := NewListener("testsocket", "unix", nil)
	if err != nil {
		t.Errorf("error listening on unix socket (%v)", err)
	}
	l.Close()
}

// TestNewListenerTLSInfoSelfCert tests that a new certificate accepts connections.
func TestNewListenerTLSInfoSelfCert(t *testing.T) {
	tmpdir, err := ioutil.TempDir(os.TempDir(), "tlsdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	tlsinfo, err := SelfCert(tmpdir, []string{"127.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	if tlsinfo.Empty() {
		t.Fatalf("tlsinfo should have certs (%+v)", tlsinfo)
	}
	testNewListenerTLSInfoAccept(t, tlsinfo)
}

func TestIsClosedConnError(t *testing.T) {
	l, err := NewListener("testsocket", "unix", nil)
	if err != nil {
		t.Errorf("error listening on unix socket (%v)", err)
	}
	l.Close()
	_, err = l.Accept()
	if !IsClosedConnError(err) {
		t.Fatalf("expect true, got false (%v)", err)
	}
}
