package groupevents

import (
	"context"
	"sgf-meetup-api/pkg/shared/models"
	"time"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

type PaginatedEventsFilters struct {
	Before *time.Time
	After  *time.Time
	Cursor string
	Limit  *int
}

func (s *Service) PaginatedEvents(ctx context.Context, groupID string, filters PaginatedEventsFilters) ([]models.MeetupEvent, *PaginatedEventsFilters, error) {
	return nil, nil, nil
}
