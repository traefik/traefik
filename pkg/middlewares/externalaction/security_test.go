package externalaction

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestSecurityMySQLConstructorsRejectUnsafeTables(t *testing.T) {
	unsafe := []string{"events` SET x=1 --", "events;DROP TABLE users", "db.events", "events/*x*/", "events\x00", "ev\nents", "events'", "events\"", "events\\", "../events", "{{ .Body.table }}", "évents", strings.Repeat("a", 65)}
	for _, table := range unsafe {
		for _, runtimeValue := range []bool{false, true} {
			for _, autoCreate := range []bool{false, true} {
				name := table
				config := dynamic.ExternalAction{}
				if runtimeValue {
					name = "runtime://table"
					config.RuntimeValues = map[string]dynamic.RuntimeValue{"table": {Value: table}}
				}
				config.Rules = []dynamic.ExternalActionRule{{
					Deduplication: &dynamic.ExternalActionDeduplication{MySQL: &dynamic.ExternalActionDedupMySQL{DSN: "unused:unused@tcp(127.0.0.1:1)/unused", Table: name, AutoCreate: autoCreate}},
					Binding:       &dynamic.ExternalActionBinding{MySQL: &dynamic.ExternalActionBindingMySQL{DSN: "unused:unused@tcp(127.0.0.1:1)/unused", Table: name, AutoCreate: autoCreate}},
				}}
				require.NoError(t, resolveRuntimeConfig(&config))
				_, err := getMySQLDeduplicator(t.Context(), *config.Rules[0].Deduplication.MySQL)
				require.ErrorContains(t, err, "mysql table must match")
				_, err = newMySQLLifecycleBindingStore(t.Context(), *config.Rules[0].Binding.MySQL)
				require.ErrorContains(t, err, "mysql table must match")
			}
		}
	}
}

func TestSQLBoundariesRejectUnsafeTable(t *testing.T) {
	table := "events`;DROP TABLE users;--"
	_, err := compileMySQLBindingQueries(table)
	require.Error(t, err)
	binding := &mysqlLifecycleBindingStore{}
	require.Error(t, binding.Bind(t.Context(), "key", "set", [32]byte{}, 0))
	_, _, err = binding.Resolve(t.Context(), "key")
	require.Error(t, err)
	binding.cleanupExpired(t.Context(), nil)
	require.Zero(t, binding.cleanup.Load())
	require.Error(t, ensureBindingSchema(t.Context(), nil, table))

	_, err = compileMySQLDedupQueries(table)
	require.Error(t, err)
	dedup := &mysqlDeduplicator{}
	_, _, err = dedup.Claim(t.Context(), "rule", "key", 0, 0)
	require.Error(t, err)
	require.Error(t, dedup.Complete(t.Context(), "key", "claim", 0))
	require.Error(t, dedup.Release(t.Context(), "key", "claim"))
	dedup.cleanupExpired(t.Context(), nil)
	require.Zero(t, dedup.cleanup.Load())
	require.Error(t, ensureDeduplicationSchema(t.Context(), nil, table))
}

func TestTransportReadsProjectedTLSFiles(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer server.Close()
	certificate := server.TLS.Certificates[0]
	key, err := x509.MarshalPKCS8PrivateKey(certificate.PrivateKey)
	require.NoError(t, err)
	directory := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(directory, "revision"), 0o700))
	require.NoError(t, os.Symlink("revision", filepath.Join(directory, "..data")))
	for name, material := range map[string][]byte{
		"cert": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Certificate[0]}),
		"key":  pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}),
	} {
		require.NoError(t, os.WriteFile(filepath.Join(directory, "revision", name), material, 0o600))
		require.NoError(t, os.Symlink(filepath.Join("..data", name), filepath.Join(directory, name)))
	}
	transport, err := buildTransport(&dynamic.ExternalActionTLS{
		CAFile: filepath.Join(directory, "cert"), CertFile: filepath.Join(directory, "cert"), KeyFile: filepath.Join(directory, "key"),
	})
	require.NoError(t, err)
	defer transport.CloseIdleConnections()
	require.Len(t, transport.TLSClientConfig.Certificates, 1)
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	response, err := transport.RoundTrip(request)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func TestTransportVerificationRequiresExplicitOptOut(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer server.Close()
	ca := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(ca, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.TLS.Certificates[0].Certificate[0]}), 0o600))
	for _, test := range []struct {
		name   string
		config *dynamic.ExternalActionTLS
		valid  bool
	}{
		{name: "default rejects untrusted"},
		{name: "empty configuration rejects untrusted", config: &dynamic.ExternalActionTLS{}},
		{name: "trusted CA", config: &dynamic.ExternalActionTLS{CAFile: ca}, valid: true},
		{name: "wrong server identity", config: &dynamic.ExternalActionTLS{CAFile: ca, ServerName: "wrong.example"}},
		{name: "explicit operator opt out", config: &dynamic.ExternalActionTLS{InsecureSkipVerify: true}, valid: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			transport, err := buildTransport(test.config)
			require.NoError(t, err)
			defer transport.CloseIdleConnections()
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
			require.NoError(t, err)
			response, err := transport.RoundTrip(request)
			if !test.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
		})
	}
}

func TestExecutableCommandRemainsOperatorSelected(t *testing.T) {
	for _, command := range []string{"sh", "{{ .Body.command }}", "/bin/echo\x00", "/bin/echo\n"} {
		_, err := newExecTarget(dynamic.ExternalActionRule{Exec: &dynamic.ExternalActionExec{Command: command}})
		require.Error(t, err)
	}
	target, err := newExecTarget(dynamic.ExternalActionRule{Exec: &dynamic.ExternalActionExec{Command: os.Args[0]}})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, target.Run(ctx, &actionData{}), context.Canceled)
}
