package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otelsdk "go.opentelemetry.io/otel/sdk/log"
)

type recordingProcessor struct {
	records []otelsdk.Record
}

func (p *recordingProcessor) Enabled(context.Context, otelsdk.EnabledParameters) bool { return true }

func (p *recordingProcessor) OnEmit(_ context.Context, record *otelsdk.Record) error {
	p.records = append(p.records, record.Clone())
	return nil
}

func (p *recordingProcessor) Shutdown(context.Context) error { return nil }

func (p *recordingProcessor) ForceFlush(context.Context) error { return nil }

func TestSamplingProcessor(t *testing.T) {
	tests := []struct {
		desc       string
		sampleRate float64
		wantCount  int
	}{
		{
			desc:       "sample rate zero drops all records",
			sampleRate: 0,
			wantCount:  0,
		},
		{
			desc:       "sample rate one keeps all records",
			sampleRate: 1,
			wantCount:  10,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := &recordingProcessor{}
			processor := newSamplingProcessor(next, test.sampleRate)

			for range 10 {
				require.NoError(t, processor.OnEmit(t.Context(), &otelsdk.Record{}))
			}

			assert.Len(t, next.records, test.wantCount)
		})
	}
}

func TestAccessLogSetDefaults(t *testing.T) {
	config := &AccessLog{}
	config.SetDefaults()

	assert.InEpsilon(t, 1.0, config.SampleRate, 0)
}
