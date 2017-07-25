package level_test

import (
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/experimental_level"
)

func BenchmarkNopBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewNopLogger())
}

func BenchmarkNopDisallowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewNopLogger(), level.Config{
		Allowed: level.AllowInfoAndAbove(),
	}))
}

func BenchmarkNopAllowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewNopLogger(), level.Config{
		Allowed: level.AllowAll(),
	}))
}

func BenchmarkJSONBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard))
}

func BenchmarkJSONDisallowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewJSONLogger(ioutil.Discard), level.Config{
		Allowed: level.AllowInfoAndAbove(),
	}))
}

func BenchmarkJSONAllowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewJSONLogger(ioutil.Discard), level.Config{
		Allowed: level.AllowAll(),
	}))
}

func BenchmarkLogfmtBaseline(b *testing.B) {
	benchmarkRunner(b, log.NewLogfmtLogger(ioutil.Discard))
}

func BenchmarkLogfmtDisallowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewLogfmtLogger(ioutil.Discard), level.Config{
		Allowed: level.AllowInfoAndAbove(),
	}))
}

func BenchmarkLogfmtAllowedLevel(b *testing.B) {
	benchmarkRunner(b, level.New(log.NewLogfmtLogger(ioutil.Discard), level.Config{
		Allowed: level.AllowAll(),
	}))
}

func benchmarkRunner(b *testing.B, logger log.Logger) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		level.Debug(logger).Log("foo", "bar")
	}
}
