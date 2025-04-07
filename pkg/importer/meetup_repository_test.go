package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"log/slog"
	"sgf-meetup-api/pkg/logging"
	"sgf-meetup-api/pkg/models"
	"testing"
	"time"
)

func TestMeetupRepository_GetEventsUntilDateForGroup(t *testing.T) {
	now := time.Now()
	faker := gofakeit.New(0)

	t.Run("single page of events before cutoff date", func(t *testing.T) {
		mock := mockPaginationHandler(faker, [][]time.Duration{
			{24 * time.Hour, 48 * time.Hour},
		})

		repo := NewMeetupRepository(mock, slog.New(logging.NewMockHandler()))
		beforeDate := now.Add(72 * time.Hour)

		events, err := repo.GetEventsUntilDateForGroup(context.Background(), "group", beforeDate)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(events) != 2 {
			t.Errorf("expected 2 events, got %d", len(events))
		}

		if mock.callCount != 1 {
			t.Errorf("expected 1 API call, got %d", mock.callCount)
		}
	})

	t.Run("multiple pages until exceeding cutoff", func(t *testing.T) {
		mock := mockPaginationHandler(faker, [][]time.Duration{
			{24 * time.Hour, 48 * time.Hour},
			{72 * time.Hour, 96 * time.Hour},
		})

		repo := NewMeetupRepository(mock, slog.New(logging.NewMockHandler()))
		beforeDate := now.Add(60 * time.Hour)

		events, err := repo.GetEventsUntilDateForGroup(context.Background(), "group", beforeDate)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(events) != 4 {
			t.Errorf("expected 4 events, got %d", len(events))
		}

		if mock.callCount != 2 {
			t.Errorf("expected 2 API calls, got %d", mock.callCount)
		}
	})

	t.Run("process all pages when no dates exceed", func(t *testing.T) {
		mock := mockPaginationHandler(faker, [][]time.Duration{
			{24 * time.Hour, 48 * time.Hour},
			{60 * time.Hour, 72 * time.Hour},
		})

		repo := NewMeetupRepository(mock, slog.New(logging.NewMockHandler()))
		beforeDate := now.Add(100 * time.Hour)

		events, err := repo.GetEventsUntilDateForGroup(context.Background(), "group", beforeDate)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(events) != 4 {
			t.Errorf("expected 4 events, got %d", len(events))
		}

		if mock.callCount != 2 {
			t.Errorf("expected 2 API calls, got %d", mock.callCount)
		}
	})

	t.Run("propagate errors from handler", func(t *testing.T) {
		failingMock := &mockGraphQLHandler{
			handlers: []func() (*MeetupFutureEventsResponse, error){
				func() (*MeetupFutureEventsResponse, error) {
					return nil, fmt.Errorf("API unavailable")
				},
			},
		}

		repo := NewMeetupRepository(failingMock, slog.New(logging.NewMockHandler()))
		_, err := repo.GetEventsUntilDateForGroup(context.Background(), "group", now)

		if err == nil {
			t.Fatal("expected error but got nil")
		}

		if failingMock.callCount != 1 {
			t.Errorf("expected 1 API call attempt, got %d", failingMock.callCount)
		}
	})
}

type mockGraphQLHandler struct {
	callCount int
	handlers  []func() (*MeetupFutureEventsResponse, error)
}

func (m *mockGraphQLHandler) ExecuteQuery(
	_ context.Context,
	_ string,
	_ map[string]any,
) ([]byte, error) {
	defer func() { m.callCount++ }()

	handler := m.handlers[m.callCount]

	resp, err := handler()

	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return data, nil
}

func generateEvents(faker *gofakeit.Faker, base time.Time, dates ...time.Duration) []models.MeetupEvent {
	events := make([]models.MeetupEvent, len(dates))
	faker.Slice(&events)
	for i, d := range dates {
		date := base.Add(d)
		events[i].DateTime = &date
	}
	return events
}

func mockPaginationHandler(faker *gofakeit.Faker, pages [][]time.Duration) *mockGraphQLHandler {
	var handlers []func() (*MeetupFutureEventsResponse, error)

	for i, offsets := range pages {
		cursor := ""
		if i < len(pages)-1 {
			cursor = uuid.New().String()
		}

		finalCursor := cursor
		events := generateEvents(faker, time.Now(), offsets...)

		handlers = append(handlers, func() (*MeetupFutureEventsResponse, error) {
			return generateMeetupResponse(events, finalCursor), nil
		})
	}

	return &mockGraphQLHandler{handlers: handlers}
}

func generateMeetupResponse(events []models.MeetupEvent, cursor string) *MeetupFutureEventsResponse {
	response := &MeetupFutureEventsResponse{}
	response.Data.Events.UnifiedEvents.Count = len(events)
	response.Data.Events.UnifiedEvents.PageInfo.EndCursor = cursor
	response.Data.Events.UnifiedEvents.PageInfo.HasNextPage = cursor != ""

	edges := make([]MeetupEdge, 0, len(events))

	for _, event := range events {
		edges = append(edges, MeetupEdge{Node: event})
	}

	response.Data.Events.UnifiedEvents.Edges = edges
	return response
}
