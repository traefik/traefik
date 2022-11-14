package safe

import "testing"

func TestSafe(t *testing.T) {
	const ts1 = "test1"
	const ts2 = "test2"

	s := NewSync(ts1)

	if result := s.Get(); result != ts1 {
		t.Errorf("Safe.Get() failed, got '%s', expected '%s'", result, ts1)
	}

	s.Set(ts2)

	if result := s.Get(); result != ts2 {
		t.Errorf("Safe.Get() after Safe.Set() failed, got '%s', expected '%s'", result, ts2)
	}
}
