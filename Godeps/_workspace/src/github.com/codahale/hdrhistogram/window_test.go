package hdrhistogram_test

import (
	"testing"

	"github.com/codahale/hdrhistogram"
)

func TestWindowedHistogram(t *testing.T) {
	w := hdrhistogram.NewWindowed(2, 1, 1000, 3)

	for i := 0; i < 100; i++ {
		w.Current.RecordValue(int64(i))
	}
	w.Rotate()

	for i := 100; i < 200; i++ {
		w.Current.RecordValue(int64(i))
	}
	w.Rotate()

	for i := 200; i < 300; i++ {
		w.Current.RecordValue(int64(i))
	}

	if v, want := w.Merge().ValueAtQuantile(50), int64(199); v != want {
		t.Errorf("Median was %v, but expected %v", v, want)
	}
}

func BenchmarkWindowedHistogramRecordAndRotate(b *testing.B) {
	w := hdrhistogram.NewWindowed(3, 1, 10000000, 3)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := w.Current.RecordValue(100); err != nil {
			b.Fatal(err)
		}

		if i%100000 == 1 {
			w.Rotate()
		}
	}
}

func BenchmarkWindowedHistogramMerge(b *testing.B) {
	w := hdrhistogram.NewWindowed(3, 1, 10000000, 3)
	for i := 0; i < 10000000; i++ {
		if err := w.Current.RecordValue(100); err != nil {
			b.Fatal(err)
		}

		if i%100000 == 1 {
			w.Rotate()
		}
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.Merge()
	}
}
