package clock

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMockTimeControl(t *testing.T) {
	initial := time.Date(2025, 4, 6, 2, 0, 0, 0, time.UTC)
	mock := MockTimeSource(initial)

	assert.Equal(t, initial, mock.Now())

	newTime := initial.Add(2 * time.Hour)
	mock.SetTime(newTime)

	assert.Equal(t, newTime, mock.Now())
}

func TestRealTimeSource(t *testing.T) {
	clock := RealTimeSource()
	before := time.Now()
	now := clock.Now()
	after := time.Now()

	assert.True(t, now.After(before))
	assert.True(t, now.Before(after))
}

func TestMockZeroTime(t *testing.T) {
	zeroTime := time.Time{}
	mock := MockTimeSource(zeroTime)

	assert.True(t, mock.Now().IsZero())

	mock.SetTime(zeroTime.Add(1 * time.Nanosecond))

	assert.False(t, mock.Now().Equal(zeroTime))
}
