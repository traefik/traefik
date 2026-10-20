package externalaction

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestEnrichmentRedactsLookupCredentialErrors(t *testing.T) {
	const secret = "lookup-only-sensitive-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, secret, r.Header.Get("X-Token"))
		_, _ = w.Write([]byte(`{"message":"` + secret + `"}`))
	}))
	t.Cleanup(server.Close)
	for _, stage := range []string{"body", "header", "mapping", "transport"} {
		t.Run(stage, func(t *testing.T) {
			config := enrichmentFixture(server.URL)
			query := &config.Queries[0]
			query.HTTP.Auth = &dynamic.ExternalActionAuth{
				Type: "header", Header: "X-Token", Token: secret,
				Parameters: map[string]string{"credential": secret},
			}
			switch stage {
			case "body":
				query.Body.Template = `{{ fail (index .AuthParameters "credential") }}`
			case "header":
				query.HTTP.Headers = map[string]string{"X-Example": `{{ fail (index .AuthParameters "credential") }}`}
			case "mapping":
				query.ResultTemplate = `{{ fail (index .Result "message") }}`
			case "transport":
				query.HTTP.URL = server.URL + `/{{ index .AuthParameters "credential" }}`
			}
			compiled, err := compileEnrichment(t.Name(), config)
			require.NoError(t, err)
			if stage == "transport" {
				compiled.queries[0].target.perform = func(r *http.Request) (*http.Response, error) {
					return nil, errors.New("request failed: " + r.URL.String())
				}
			}
			_, _, err = enrichPayload(t.Context(), &actionData{Context: t.Context()}, `{"id":1}`, []*compiledEnrichment{compiled})
			require.Error(t, err)
			assert.NotContains(t, err.Error(), secret)
			assert.Contains(t, err.Error(), "<redacted>")
		})
	}
}

func TestCredentialRedactionPreservesNonSecretErrors(t *testing.T) {
	auth := &boundAuth{secrets: []string{"sensitive-token"}}
	require.NoError(t, auth.redactError(nil))
	require.ErrorIs(t, auth.redactError(context.Canceled), context.Canceled)
	require.ErrorIs(t, auth.redactError(context.DeadlineExceeded), context.DeadlineExceeded)
}
