package hdrhistogram_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/codahale/hdrhistogram"
)

func TestHighSigFig(t *testing.T) {
	input := []int64{
		459876, 669187, 711612, 816326, 931423, 1033197, 1131895, 2477317,
		3964974, 12718782,
	}

	hist := hdrhistogram.New(459876, 12718782, 5)
	for _, sample := range input {
		hist.RecordValue(sample)
	}

	if v, want := hist.ValueAtQuantile(50), int64(1048575); v != want {
		t.Errorf("Median was %v, but expected %v", v, want)
	}
}

func TestValueAtQuantile(t *testing.T) {
	h := hdrhistogram.New(1, 10000000, 3)

	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	data := []struct {
		q float64
		v int64
	}{
		{q: 50, v: 500223},
		{q: 75, v: 750079},
		{q: 90, v: 900095},
		{q: 95, v: 950271},
		{q: 99, v: 990207},
		{q: 99.9, v: 999423},
		{q: 99.99, v: 999935},
	}

	for _, d := range data {
		if v := h.ValueAtQuantile(d.q); v != d.v {
			t.Errorf("P%v was %v, but expected %v", d.q, v, d.v)
		}
	}
}

func TestMean(t *testing.T) {
	h := hdrhistogram.New(1, 10000000, 3)

	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	if v, want := h.Mean(), 500000.013312; v != want {
		t.Errorf("Mean was %v, but expected %v", v, want)
	}
}

func TestStdDev(t *testing.T) {
	h := hdrhistogram.New(1, 10000000, 3)

	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	if v, want := h.StdDev(), 288675.1403682715; v != want {
		t.Errorf("StdDev was %v, but expected %v", v, want)
	}
}

func TestTotalCount(t *testing.T) {
	h := hdrhistogram.New(1, 10000000, 3)

	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
		if v, want := h.TotalCount(), int64(i+1); v != want {
			t.Errorf("TotalCount was %v, but expected %v", v, want)
		}
	}
}

func TestMax(t *testing.T) {
	h := hdrhistogram.New(1, 10000000, 3)

	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	if v, want := h.Max(), int64(999936); v != want {
		t.Errorf("Max was %v, but expected %v", v, want)
	}
}

func TestReset(t *testing.T) {
	h := hdrhistogram.New(1, 10000000, 3)

	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	h.Reset()

	if v, want := h.Max(), int64(0); v != want {
		t.Errorf("Max was %v, but expected %v", v, want)
	}
}

func TestMerge(t *testing.T) {
	h1 := hdrhistogram.New(1, 1000, 3)
	h2 := hdrhistogram.New(1, 1000, 3)

	for i := 0; i < 100; i++ {
		if err := h1.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	for i := 100; i < 200; i++ {
		if err := h2.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	h1.Merge(h2)

	if v, want := h1.ValueAtQuantile(50), int64(99); v != want {
		t.Errorf("Median was %v, but expected %v", v, want)
	}
}

func TestMin(t *testing.T) {
	h := hdrhistogram.New(1, 10000000, 3)

	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	if v, want := h.Min(), int64(0); v != want {
		t.Errorf("Min was %v, but expected %v", v, want)
	}
}

func TestByteSize(t *testing.T) {
	h := hdrhistogram.New(1, 100000, 3)

	if v, want := h.ByteSize(), 65604; v != want {
		t.Errorf("ByteSize was %v, but expected %d", v, want)
	}
}

func TestRecordCorrectedValue(t *testing.T) {
	h := hdrhistogram.New(1, 100000, 3)

	if err := h.RecordCorrectedValue(10, 100); err != nil {
		t.Fatal(err)
	}

	if v, want := h.ValueAtQuantile(75), int64(10); v != want {
		t.Errorf("Corrected value was %v, but expected %v", v, want)
	}
}

func TestRecordCorrectedValueStall(t *testing.T) {
	h := hdrhistogram.New(1, 100000, 3)

	if err := h.RecordCorrectedValue(1000, 100); err != nil {
		t.Fatal(err)
	}

	if v, want := h.ValueAtQuantile(75), int64(800); v != want {
		t.Errorf("Corrected value was %v, but expected %v", v, want)
	}
}

func TestCumulativeDistribution(t *testing.T) {
	h := hdrhistogram.New(1, 100000000, 3)

	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	actual := h.CumulativeDistribution()
	expected := []hdrhistogram.Bracket{
		hdrhistogram.Bracket{Quantile: 0, Count: 1, ValueAt: 0},
		hdrhistogram.Bracket{Quantile: 50, Count: 500224, ValueAt: 500223},
		hdrhistogram.Bracket{Quantile: 75, Count: 750080, ValueAt: 750079},
		hdrhistogram.Bracket{Quantile: 87.5, Count: 875008, ValueAt: 875007},
		hdrhistogram.Bracket{Quantile: 93.75, Count: 937984, ValueAt: 937983},
		hdrhistogram.Bracket{Quantile: 96.875, Count: 969216, ValueAt: 969215},
		hdrhistogram.Bracket{Quantile: 98.4375, Count: 984576, ValueAt: 984575},
		hdrhistogram.Bracket{Quantile: 99.21875, Count: 992256, ValueAt: 992255},
		hdrhistogram.Bracket{Quantile: 99.609375, Count: 996352, ValueAt: 996351},
		hdrhistogram.Bracket{Quantile: 99.8046875, Count: 998400, ValueAt: 998399},
		hdrhistogram.Bracket{Quantile: 99.90234375, Count: 999424, ValueAt: 999423},
		hdrhistogram.Bracket{Quantile: 99.951171875, Count: 999936, ValueAt: 999935},
		hdrhistogram.Bracket{Quantile: 99.9755859375, Count: 999936, ValueAt: 999935},
		hdrhistogram.Bracket{Quantile: 99.98779296875, Count: 999936, ValueAt: 999935},
		hdrhistogram.Bracket{Quantile: 99.993896484375, Count: 1000000, ValueAt: 1000447},
		hdrhistogram.Bracket{Quantile: 100, Count: 1000000, ValueAt: 1000447},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("CF was %#v, but expected %#v", actual, expected)
	}
}

func TestNaN(t *testing.T) {
	h := hdrhistogram.New(1, 100000, 3)
	if math.IsNaN(h.Mean()) {
		t.Error("mean is NaN")
	}
	if math.IsNaN(h.StdDev()) {
		t.Error("stddev is NaN")
	}
}

func BenchmarkHistogramRecordValue(b *testing.B) {
	h := hdrhistogram.New(1, 10000000, 3)
	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		h.RecordValue(100)
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		hdrhistogram.New(1, 120000, 3) // this could track 1ms-2min
	}
}

func TestUnitMagnitudeOverflow(t *testing.T) {
	h := hdrhistogram.New(0, 200, 4)
	if err := h.RecordValue(11); err != nil {
		t.Fatal(err)
	}
}

func TestSubBucketMaskOverflow(t *testing.T) {
	hist := hdrhistogram.New(2e7, 1e8, 5)
	for _, sample := range [...]int64{1e8, 2e7, 3e7} {
		hist.RecordValue(sample)
	}

	for q, want := range map[float64]int64{
		50:    33554431,
		83.33: 33554431,
		83.34: 100663295,
		99:    100663295,
	} {
		if got := hist.ValueAtQuantile(q); got != want {
			t.Errorf("got %d for %fth percentile. want: %d", got, q, want)
		}
	}
}

func TestExportImport(t *testing.T) {
	min := int64(1)
	max := int64(10000000)
	sigfigs := 3
	h := hdrhistogram.New(min, max, sigfigs)
	for i := 0; i < 1000000; i++ {
		if err := h.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	s := h.Export()

	if v := s.LowestTrackableValue; v != min {
		t.Errorf("LowestTrackableValue was %v, but expected %v", v, min)
	}

	if v := s.HighestTrackableValue; v != max {
		t.Errorf("HighestTrackableValue was %v, but expected %v", v, max)
	}

	if v := int(s.SignificantFigures); v != sigfigs {
		t.Errorf("SignificantFigures was %v, but expected %v", v, sigfigs)
	}

	if imported := hdrhistogram.Import(s); !imported.Equals(h) {
		t.Error("Expected Histograms to be equivalent")
	}

}

func TestEquals(t *testing.T) {
	h1 := hdrhistogram.New(1, 10000000, 3)
	for i := 0; i < 1000000; i++ {
		if err := h1.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	h2 := hdrhistogram.New(1, 10000000, 3)
	for i := 0; i < 10000; i++ {
		if err := h1.RecordValue(int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	if h1.Equals(h2) {
		t.Error("Expected Histograms to not be equivalent")
	}

	h1.Reset()
	h2.Reset()

	if !h1.Equals(h2) {
		t.Error("Expected Histograms to be equivalent")
	}
}
