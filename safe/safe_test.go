package safe

import "testing"

func TestSafe(t *testing.T) {
	const ts1 = "test1"
	const ts2 = "test2"

	s := New(ts1)

	result, ok := s.Get().(string)
	if !ok {
		t.Fatalf("Safe.Get() failed, got type '%T', expected string", s.Get())
	}

	if result != ts1 {
		t.Errorf("Safe.Get() failed, got '%s', expected '%s'", result, ts1)
	}

	s.Set(ts2)

	result, ok = s.Get().(string)
	if !ok {
		t.Fatalf("Safe.Get() after Safe.Set() failed, got type '%T', expected string", s.Get())
	}

	if result != ts2 {
		t.Errorf("Safe.Get() after Safe.Set() failed, got '%s', expected '%s'", result, ts2)
	}
}
