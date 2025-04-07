package clock

import "time"

type TimeSource interface {
	Now() time.Time
}

type TimeControl interface {
	TimeSource
	SetTime(time.Time)
}

type realTimeSource struct{}

func (r realTimeSource) Now() time.Time {
	return time.Now()
}

func RealTimeSource() TimeSource {
	return &realTimeSource{}
}

type mockTimeSource struct {
	frozenTime time.Time
}

func (f *mockTimeSource) Now() time.Time {
	return f.frozenTime
}

func (f *mockTimeSource) SetTime(t time.Time) {
	f.frozenTime = t
}

func MockTimeSource(initialTime time.Time) TimeControl {
	return &mockTimeSource{frozenTime: initialTime}
}
