package timetools

import (
	"fmt"
	"testing"
	"time"
)

var _ = fmt.Printf // for testing

func TestRealTimeUtcNow(t *testing.T) {
	rt := RealTime{}

	rtNow := rt.UtcNow()
	atNow := time.Now().UTC()

	// times shouldn't be exact
	if rtNow.Equal(atNow) {
		t.Errorf("rt.UtcNow() = time.Now.UTC(), %v = %v, should be slightly different", rtNow, atNow)
	}

	rtNowPlusOne := atNow.Add(1 * time.Second)
	rtNowMinusOne := atNow.Add(-1 * time.Second)

	// but should be pretty close
	if atNow.After(rtNowPlusOne) || atNow.Before(rtNowMinusOne) {
		t.Errorf("timedelta between rt.UtcNow() and time.Now.UTC() greater than 2 seconds, %v, %v", rtNow, atNow)
	}
}

func TestFreezeTimeUtcNow(t *testing.T) {
	tm := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	ft := FreezedTime{tm}

	if !tm.Equal(ft.UtcNow()) {
		t.Errorf("ft.UtcNow() != time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), %v, %v", tm, ft)
	}
}

func itTicks(c <-chan time.Time) bool {
	select {
	case <-c:
		return true
	case <-time.After(time.Millisecond):
		return false
	}
}

func TestSleepableTime(t *testing.T) {
	tm := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	st := SleepProvider(tm)

	if !tm.Equal(st.UtcNow()) {
		t.Errorf("st.UtcNow() != time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), %v, %v", tm, st)
	}

	// Check After with no AdvanceTimeBy
	if itTicks(st.After(time.Nanosecond)) {
		t.Error("Got tick from After before calling AdvanceTimeBy")
	}

	// Check After with one call to AdvanceTimeBy
	c0 := st.After(time.Hour)
	AdvanceTimeBy(st, 2*time.Hour)
	if !itTicks(c0) {
		t.Error("Didn't get tick from After after calling AdvanceTimeBy")
	}

	// Check After with multiple calls to AdvanceTimeBy
	c0 = st.After(time.Hour)
	AdvanceTimeBy(st, 20*time.Minute)
	if itTicks(c0) {
		t.Error("Got tick from After before we AdvanceTimeBy'd enough")
	}
	AdvanceTimeBy(st, 20*time.Minute)
	if itTicks(c0) {
		t.Error("Got tick from After before we AdvanceTimeBy'd enough")
	}
	AdvanceTimeBy(st, 40*time.Minute)
	if !itTicks(c0) {
		t.Error("Didn't get tick from After after we AdvanceTimeBy'd enough")
	}

	// Check Sleep with no AdvanceTimeBy
	c1 := make(chan time.Time)
	go func() {
		st.Sleep(time.Nanosecond)
		c1 <- st.UtcNow()
	}()
	if itTicks(c1) {
		t.Error("Sleep returned before we called AdvanceTimeBy")
	}
}
