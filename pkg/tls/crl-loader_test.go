package tls

import (
	"bytes"
	"crypto/x509"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrlNOOPLoader_Load(t *testing.T) {
	loader := crlNOOPLoader{}
	snap, err := loader.Load(&crlEntry{})
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0), snap.crlSerial())
	assert.Equal(t, 0, snap.serialCount())
	assert.False(t, snap.needsRefresh(time.Second))
}

func TestCrlFileLoader_Load_Success(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 7, []x509.RevocationListEntry{
		{SerialNumber: big.NewInt(123), RevocationTime: time.Now().Add(-time.Hour)},
	}, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	path := filepath.Join(t.TempDir(), "test.crl")
	require.NoError(t, os.WriteFile(path, der, 0o600))

	loader := &crlFileLoader{path: path}
	snap, err := loader.Load(&crlEntry{})
	require.NoError(t, err)

	assert.Equal(t, big.NewInt(7), snap.crlSerial())
	assert.True(t, snap.containsSerial(big.NewInt(123)))
	assert.False(t, snap.containsSerial(big.NewInt(1)))
}

func TestCrlFileLoader_Load_MissingFile(t *testing.T) {
	loader := &crlFileLoader{path: "/does/not/exist.crl"}
	_, err := loader.Load(&crlEntry{})
	assert.Error(t, err)
}

func TestCrlFileLoader_Load_InvalidContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.crl")
	require.NoError(t, os.WriteFile(path, []byte("garbage"), 0o600))

	loader := &crlFileLoader{path: path}
	_, err := loader.Load(&crlEntry{})
	assert.Error(t, err)
}

func TestCrlHTTPLoader_Load_Success(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, []x509.RevocationListEntry{
		{SerialNumber: big.NewInt(55), RevocationTime: time.Now()},
	}, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(der)
	})

	loader := &crlHTTPLoader{
		distributionPoint: srv.URL,

		clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024),
	}

	entry := &crlEntry{}
	entry.issuerPtr.Store(ca.cert)
	snap, err := loader.Load(entry)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(1), snap.crlSerial())
	assert.True(t, snap.containsSerial(big.NewInt(55)))
}

func TestCrlHTTPLoader_Load_HTTPError(t *testing.T) {
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	loader := &crlHTTPLoader{distributionPoint: srv.URL, clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := loader.Load(&crlEntry{})
	assert.Error(t, err)
}

func TestCrlHTTPLoader_Load_InvalidBody(t *testing.T) {
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not a crl"))
	})

	loader := &crlHTTPLoader{distributionPoint: srv.URL, clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := loader.Load(&crlEntry{})
	assert.Error(t, err)
}

func TestCrlHTTPLoader_Load_SignatureMismatch(t *testing.T) {
	legitimateCA := newTestCA(t)
	rogueCA := newTestCA(t)

	// CRL signed by a different CA than the configured issuer.
	der := newTestCRLDER(t, rogueCA, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	loader := &crlHTTPLoader{distributionPoint: srv.URL, clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := &crlEntry{}
	entry.issuerPtr.Store(legitimateCA.cert)
	_, err := loader.Load(entry)
	assert.Error(t, err)
}

func TestCrlHTTPLoader_Load_Expired(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	loader := &crlHTTPLoader{distributionPoint: srv.URL, clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := &crlEntry{}
	entry.issuerPtr.Store(ca.cert)
	_, err := loader.Load(entry)
	assert.Error(t, err)
}

func TestCrlHTTPLoader_Load_ThisUpdateInFuture(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(time.Hour), time.Now().Add(2*time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	loader := &crlHTTPLoader{distributionPoint: srv.URL, clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := &crlEntry{}
	entry.issuerPtr.Store(ca.cert)
	_, err := loader.Load(entry)
	assert.Error(t, err)
}

func TestCrlHTTPLoader_Load_RollbackDetected(t *testing.T) {
	ca := newTestCA(t)

	entry := &crlEntry{}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{number: big.NewInt(10)}})
	entry.issuerPtr.Store(ca.cert)

	der := newTestCRLDER(t, ca, 5, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	loader := &crlHTTPLoader{distributionPoint: srv.URL, clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := loader.Load(entry)
	assert.ErrorContains(t, err, "rollback")
}

func TestCrlHTTPLoader_Load_NoRollbackOnEqualOrHigherNumber(t *testing.T) {
	ca := newTestCA(t)

	entry := &crlEntry{}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{number: big.NewInt(5)}})
	entry.issuerPtr.Store(ca.cert)

	der := newTestCRLDER(t, ca, 6, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	loader := &crlHTTPLoader{distributionPoint: srv.URL, clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	snap, err := loader.Load(entry)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(6), snap.crlSerial())
}

func TestCrlHTTPLoader_VerifyCRLFreshness_Table(t *testing.T) {
	loader := &crlHTTPLoader{}
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name                   string
		thisUpdate, nextUpdate time.Time
		wantErr                bool
	}{
		{"valid window", now.Add(-time.Hour), now.Add(time.Hour), false},
		{"thisUpdate in future", now.Add(time.Hour), now.Add(2 * time.Hour), true},
		{"nextUpdate passed", now.Add(-2 * time.Hour), now.Add(-time.Hour), true},
		{"zero nextUpdate ignored", now.Add(-time.Hour), time.Time{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crl := &x509.RevocationList{ThisUpdate: tt.thisUpdate, NextUpdate: tt.nextUpdate}
			err := loader.verifyCRLFreshness(crl, now)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCrlHTTPLoader_VerifyCRLNumberMonotonic_Table(t *testing.T) {
	loader := &crlHTTPLoader{}
	newSnap := func(n int64) crlSnapshot {
		return &dynamicCrlSnapshot{crlSnapshotCommon{number: big.NewInt(n)}}
	}

	tests := []struct {
		name           string
		oldCRL, newCRL crlSnapshot
		wantErr        bool
	}{
		{"nil old snapshot", nil, newSnap(1), false},
		{"old without serial", &dynamicCrlSnapshot{}, newSnap(1), false},
		{"new without serial", newSnap(1), &dynamicCrlSnapshot{}, false},
		{"increasing number", newSnap(1), newSnap(2), false},
		{"equal number", newSnap(2), newSnap(2), false},
		{"decreasing number", newSnap(5), newSnap(2), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.verifyCRLNumberMonotonic(tt.newCRL, tt.oldCRL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCrlHTTPClientHolder_SetGet(t *testing.T) {
	holder := newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)
	client1 := holder.get()
	require.NotNil(t, client1)
	assert.Equal(t, time.Second, client1.client.Timeout)

	holder.set(nil, 5*time.Second, 1024)
	client2 := holder.get()
	assert.Equal(t, 5*time.Second, client2.client.Timeout, 1024)
	assert.NotSame(t, client1, client2)
}

func TestCrlHTTPClientHolder_ConcurrentAccess(t *testing.T) {
	holder := newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			holder.set(nil, time.Duration(i)*time.Millisecond, 1024)
		}
		close(done)
	}()

	for i := 0; i < 1000; i++ {
		_ = holder.get()
	}
	<-done
}

func TestCrlHTTPLoader_DownloadCRL_SizeLimit_Table(t *testing.T) {
	tests := []struct {
		name        string
		bodySize    int
		maxCRLBytes int64
		wantErr     bool
	}{
		{"body smaller than max", 100, 200, false},
		{"body equal to max", 200, 200, false},
		{"body larger than max", 201, 200, true},
		{"empty body, zero max", 0, 0, false},
		{"nonempty body, zero max", 1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.Repeat([]byte("A"), tt.bodySize)
			srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(body)
			})

			loader := &crlHTTPLoader{
				distributionPoint: srv.URL,
				clientHolder:      newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, tt.maxCRLBytes),
			}

			data, err := loader.downloadCRL()
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "exceeds the maximum allowed size")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.bodySize, len(data))
			}
		})
	}
}

func TestCrlHTTPLoader_Load_CRLExceedsMaxSize(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(der)
	})

	loader := &crlHTTPLoader{
		distributionPoint: srv.URL,

		clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, int64(len(der)-1)),
	}
	entry := &crlEntry{}
	entry.issuerPtr.Store(ca.cert)
	_, err := loader.Load(entry)
	assert.ErrorContains(t, err, "exceeds the maximum allowed size")
}

func TestCrlHTTPLoader_Load_CRLAtExactMaxSize(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(der)
	})

	loader := &crlHTTPLoader{
		distributionPoint: srv.URL,

		clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, int64(len(der))),
	}

	entry := &crlEntry{}
	entry.issuerPtr.Store(ca.cert)
	snap, err := loader.Load(entry)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(1), snap.crlSerial())
}

// --- crlHTTPLoader guards against a missing issuer ---

func TestCrlHTTPLoader_Load_NoIssuerRecorded(t *testing.T) {
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("CRL should not be downloaded without a recorded issuer")
	})

	entry := &crlEntry{
		loader: &crlHTTPLoader{
			distributionPoint: srv.URL,
			clientHolder:      newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024),
		},
	}

	_, err := entry.loader.Load(entry)
	assert.Error(t, err)
}

func TestCrlHTTPLoader_Load_IssuerRotationWithLowerCRLNumber_FalsePositiveRollback(t *testing.T) {
	// KNOWN LIMITATION: anti-rollback is tracked per distribution point URL, not
	// per issuing CA. Here the cached snapshot (number=100) conceptually comes
	// from a previous CA that used to publish CRLs at this same URL. The entry's
	// issuer has since rotated to a new CA whose own CRL numbering restarts at a
	// lower value. The legitimate new CRL ends up being rejected as a rollback.
	newCA := newTestCA(t)

	entry := &crlEntry{}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{number: big.NewInt(100)}})
	entry.issuerPtr.Store(newCA.cert)

	der := newTestCRLDER(t, newCA, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	loader := &crlHTTPLoader{
		distributionPoint: srv.URL,
		clientHolder:      newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024),
	}

	_, err := loader.Load(entry)
	assert.ErrorContains(t, err, "rollback")
}
