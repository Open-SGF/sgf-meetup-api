package clock

import (
	"testing"
	"time"
)

func TestMockTimeControl(t *testing.T) {
	initial := time.Date(2025, 4, 6, 2, 0, 0, 0, time.UTC)
	mock := NewMock(initial)

	if now := mock.Now(); !now.Equal(initial) {
		t.Errorf("Expected initial time %v, got %v", initial, now)
	}

	newTime := initial.Add(2 * time.Hour)
	mock.SetTime(newTime)
	if now := mock.Now(); !now.Equal(newTime) {
		t.Errorf("Expected updated time %v, got %v", newTime, now)
	}
}

func TestRealTimeSource(t *testing.T) {
	clock := NewReal()
	before := time.Now()
	now := clock.Now()
	after := time.Now()

	if now.Before(before) || now.After(after) {
		t.Errorf("Real time %v not in expected range [%v, %v]", now, before, after)
	}
}

func TestMockZeroTime(t *testing.T) {
	zeroTime := time.Time{}
	mock := NewMock(zeroTime)

	if !mock.Now().IsZero() {
		t.Error("NewMock should handle zero time correctly")
	}

	mock.SetTime(zeroTime.Add(1 * time.Nanosecond))
	if mock.Now().Equal(zeroTime) {
		t.Error("SetTime should allow fractional nanosecond adjustments")
	}
}
