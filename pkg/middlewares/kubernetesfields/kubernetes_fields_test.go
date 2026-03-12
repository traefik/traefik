package kubernetesfields

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
)

func TestKubernetesFields_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc                 string
		kubernetesFields     dynamic.KubernetesFields
		inputLogDataTable    *accesslog.LogData
		expectedLogDataTable *accesslog.LogData
	}{
		{
			desc: "no LogDataTable",
			kubernetesFields: dynamic.KubernetesFields{
				Namespace: "foo",
				Kind:      "MyKind",
				Name:      "bar",
			},
		},
		{
			desc: "empty LogDataTable",
			kubernetesFields: dynamic.KubernetesFields{
				Namespace: "foo",
				Kind:      "MyKind",
				Name:      "bar",
			},
			inputLogDataTable: &accesslog.LogData{},
			expectedLogDataTable: &accesslog.LogData{
				Core: accesslog.CoreLogData{
					"KubernetesNamespace": "foo",
					"KubernetesKind":      "MyKind",
					"KubernetesName":      "bar",
				},
			},
		},
		{
			desc: "existing LogDataTable",
			kubernetesFields: dynamic.KubernetesFields{
				Namespace: "foo",
				Kind:      "MyKind",
				Name:      "bar",
			},
			inputLogDataTable: &accesslog.LogData{
				Core: accesslog.CoreLogData{
					"RequestAddr": "10.10.10.10:80",
				},
			},
			expectedLogDataTable: &accesslog.LogData{
				Core: accesslog.CoreLogData{
					"RequestAddr":         "10.10.10.10:80",
					"KubernetesNamespace": "foo",
					"KubernetesKind":      "MyKind",
					"KubernetesName":      "bar",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			kubernetesFieldsHandler, err := New(t.Context(), next, test.kubernetesFields, "traefikTest")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "http://10.10.10.10", nil)

			if test.inputLogDataTable != nil {
				req = req.WithContext(context.WithValue(req.Context(), accesslog.DataTableKey, test.inputLogDataTable))
			}
			kubernetesFieldsHandler.ServeHTTP(recorder, req)

			actualLogDataTable, _ := req.Context().Value(accesslog.DataTableKey).(*accesslog.LogData)
			assert.Equal(t, test.expectedLogDataTable, actualLogDataTable)
			assert.Equal(t, 200, recorder.Code)
		})
	}
}
