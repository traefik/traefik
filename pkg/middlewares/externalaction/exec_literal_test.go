package externalaction

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestExecRequestArgumentsRemainLiteral(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "must-not-exist")
	for _, payload := range []string{
		"; touch " + marker,
		"$(touch " + marker + ")",
		"`touch " + marker + "`",
		"line one\nline two",
		`"quoted" 'single' \ slash`,
		"$PATH * ? --help",
	} {
		t.Run(payload, func(t *testing.T) {
			destination, err := newExecTarget(dynamic.ExternalActionRule{Exec: &dynamic.ExternalActionExec{
				Command: os.Args[0],
				Args:    []string{"-test.run=^TestExecLiteralArgumentsHelper$", "--", `{{ index .Body "value" }}`},
				Env: map[string]string{
					"EXTERNAL_ACTION_LITERAL_PROBE": "1",
					"EXTERNAL_ACTION_EXPECTED_ARG":  payload,
				},
			}})
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			require.NoError(t, destination.Run(ctx, &actionData{Context: ctx, Body: map[string]any{"value": payload}}))
			_, err = os.Stat(marker)
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestExecLiteralArgumentsHelper(t *testing.T) {
	if os.Getenv("EXTERNAL_ACTION_LITERAL_PROBE") != "1" {
		return
	}
	index := slices.Index(os.Args, "--")
	require.NotEqual(t, -1, index)
	require.Equal(t, []string{os.Getenv("EXTERNAL_ACTION_EXPECTED_ARG")}, os.Args[index+1:])
}
