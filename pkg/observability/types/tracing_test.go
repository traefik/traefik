package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTracingVerbosity_Allows(t *testing.T) {
	tests := []struct {
		desc   string
		from   TracingVerbosity
		to     TracingVerbosity
		allows bool
	}{
		{
			desc:   "minimal vs minimal",
			from:   MinimalVerbosity,
			to:     MinimalVerbosity,
			allows: true,
		},
		{
			desc:   "minimal vs detailed",
			from:   MinimalVerbosity,
			to:     DetailedVerbosity,
			allows: false,
		},
		{
			desc:   "detailed vs minimal",
			from:   DetailedVerbosity,
			to:     MinimalVerbosity,
			allows: true,
		},
		{
			desc:   "detailed vs detailed",
			from:   DetailedVerbosity,
			to:     DetailedVerbosity,
			allows: true,
		},
		{
			desc:   "unknown vs minimal",
			from:   TracingVerbosity("unknown"),
			to:     MinimalVerbosity,
			allows: true,
		},
		{
			desc:   "unknown vs detailed",
			from:   TracingVerbosity("unknown"),
			to:     DetailedVerbosity,
			allows: false,
		},
		{
			desc:   "minimal vs unknown",
			from:   MinimalVerbosity,
			to:     TracingVerbosity("unknown"),
			allows: false,
		},
		{
			desc:   "detailed vs unknown",
			from:   DetailedVerbosity,
			to:     TracingVerbosity("unknown"),
			allows: false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.allows, test.from.Allows(test.to))
		})
	}
}
