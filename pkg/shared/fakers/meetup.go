package fakers

import (
	"github.com/brianvoe/gofakeit/v7"
	"sgf-meetup-api/pkg/shared/models"
	"time"
)

type MeetupFaker struct {
	faker *gofakeit.Faker
}

func NewMeetupFaker(seed uint64) *MeetupFaker {
	return &MeetupFaker{
		faker: gofakeit.New(seed),
	}
}

func (m *MeetupFaker) CreateEvent(groupID string, dateTime time.Time) models.MeetupEvent {
	event := models.MeetupEvent{}
	_ = m.faker.Struct(&event)
	event.GroupID = groupID
	event.DateTime = &dateTime
	return event
}

func (m *MeetupFaker) CreateEvents(groupID string, count int) []models.MeetupEvent {
	events := make([]models.MeetupEvent, count)
	m.faker.Slice(&events)
	for i := range events {
		events[i].GroupID = groupID
	}
	return events
}

func (m *MeetupFaker) CreateEventsWithDates(groupID string, base time.Time, dates ...time.Duration) []models.MeetupEvent {
	events := make([]models.MeetupEvent, len(dates))
	m.faker.Slice(&events)
	for i, d := range dates {
		date := base.Add(d)
		events[i].GroupID = groupID
		events[i].DateTime = &date
	}
	return events
}
