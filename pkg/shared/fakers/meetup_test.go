package fakers

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMeetupFaker_CreateEvent(t *testing.T) {
	faker := NewMeetupFaker(0)

	group := "group"
	now := time.Now()
	event := faker.CreateEvent(group, now)

	assert.Equal(t, group, event.GroupID)
	assert.Equal(t, now, *event.DateTime)
}

func TestMeetupFaker_CreateEvents(t *testing.T) {
	faker := NewMeetupFaker(0)

	group := "group"
	events := faker.CreateEvents(group, 10)

	assert.Equal(t, 10, len(events))
	for _, event := range events {
		assert.Equal(t, group, event.GroupID)
	}
}

func TestMeetupFaker_CreateEventsWithDates(t *testing.T) {
	faker := NewMeetupFaker(0)

	group := "group"
	base := time.Now()
	durations := []time.Duration{time.Second * 10, time.Second * 20}
	events := faker.CreateEventsWithDates(group, base, durations...)

	assert.Equal(t, 2, len(events))
	for i, event := range events {
		assert.Equal(t, group, event.GroupID)
		assert.Equal(t, base.Add(durations[i]), *event.DateTime)
	}
}
