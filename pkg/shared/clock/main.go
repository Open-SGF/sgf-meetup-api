package clock

import (
	"github.com/google/wire"
	"time"
)

type TimeSource interface {
	Now() time.Time
}

type RealTimeSource struct{}

func NewRealTimeSource() *RealTimeSource {
	return &RealTimeSource{}
}

func (r RealTimeSource) Now() time.Time {
	return time.Now()
}

type MockTimeSource struct {
	frozenTime time.Time
}

func NewMockTimeSource(initialTime time.Time) *MockTimeSource {
	return &MockTimeSource{frozenTime: initialTime}
}

func (m *MockTimeSource) Now() time.Time {
	return m.frozenTime
}

func (m *MockTimeSource) SetTime(t time.Time) {
	m.frozenTime = t
}

var RealClockProvider = wire.NewSet(wire.Bind(new(TimeSource), new(*RealTimeSource)), NewRealTimeSource)
