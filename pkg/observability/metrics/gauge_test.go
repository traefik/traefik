package metrics

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/go-kit/kit/metrics/influx"
	"github.com/go-kit/kit/metrics/statsd"
	"github.com/go-kit/log"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatefulGauge_dropLabels(t *testing.T) {
	store := statsd.New("", log.NewNopLogger())
	gauge := newStatefulGauge(store.NewGauge("gauge"), true)

	push := func() string {
		t.Helper()

		var buf bytes.Buffer
		_, err := store.WriteTo(&buf)
		require.NoError(t, err)

		return buf.String()
	}

	gauge.With("entrypoint", "a").Add(1)
	gauge.With("entrypoint", "a").Add(1)
	gauge.With("entrypoint", "b").Add(1)
	assert.Equal(t, "gauge:3.000000|g\n", push())

	gauge.With("entrypoint", "a").Add(-1)
	assert.Equal(t, "gauge:2.000000|g\n", push())

	gauge.With("entrypoint", "b").Set(5)
	gauge.With("entrypoint", "a").Add(1)
	assert.Equal(t, "gauge:6.000000|g\n", push())
}

func TestStatefulGauge_labels(t *testing.T) {
	store := influx.New(nil, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	gauge := newStatefulGauge(store.NewGauge("gauge"), false)

	push := func() []string {
		t.Helper()

		w := &pointsWriter{}
		require.NoError(t, store.WriteTo(w))

		slices.Sort(w.points)
		return w.points
	}

	gauge.With("entrypoint", "a").Add(1)
	gauge.With("entrypoint", "a").Add(1)
	gauge.With("entrypoint", "b").Add(1)
	assert.Equal(t, []string{"gauge,entrypoint=a value=2", "gauge,entrypoint=b value=1"}, push())

	gauge.With("entrypoint", "a").Add(-1)
	assert.Equal(t, []string{"gauge,entrypoint=a value=1"}, push())

	gauge.With("entrypoint", "b").Set(5)
	gauge.With("entrypoint", "b").Add(1)
	assert.Equal(t, []string{"gauge,entrypoint=b value=6"}, push())
}

type pointsWriter struct {
	points []string
}

func (w *pointsWriter) Write(bp influxdb.BatchPoints) error {
	for _, p := range bp.Points() {
		fields, err := p.Fields()
		if err != nil {
			return err
		}

		w.points = append(w.points, fmt.Sprintf("%s,entrypoint=%s value=%v", p.Name(), p.Tags()["entrypoint"], fields["value"]))
	}
	return nil
}
