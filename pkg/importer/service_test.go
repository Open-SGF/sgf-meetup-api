package importer

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/importer/importerconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/fakers"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/shared/models"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

func TestNewServiceConfig(t *testing.T) {
	cfg := &importerconfig.Config{
		MeetupGroupNames: []string{"Test1", "Test2"},
	}

	serviceConfig := NewServiceConfig(cfg)

	assert.ElementsMatch(t, cfg.MeetupGroupNames, serviceConfig.GroupNames)
}

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) GetUpcomingEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error) {
	args := m.Called(ctx, group)
	return args.Get(0).([]models.MeetupEvent), args.Error(1)
}

func (m *MockEventRepository) ArchiveEvents(ctx context.Context, eventIds []string) error {
	args := m.Called(ctx, eventIds)
	return args.Error(0)
}

func (m *MockEventRepository) UpsertEvents(ctx context.Context, events []models.MeetupEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

type MockMeetupRepository struct {
	mock.Mock
}

func (m *MockMeetupRepository) GetEventsUntilDateForGroup(ctx context.Context, group string, beforeDate time.Time) ([]models.MeetupEvent, error) {
	args := m.Called(ctx, group, beforeDate)
	return args.Get(0).([]models.MeetupEvent), args.Error(1)
}

func TestService_Import(t *testing.T) {
	ctx := context.Background()
	meetupFaker := fakers.NewMeetupFaker(0)
	now := time.Now()

	t.Run("processes all groups with concurrency control", func(t *testing.T) {
		mockTimeSource := clock.NewMockTimeSource(now)
		eventRepo := new(MockEventRepository)
		meetupRepo := new(MockMeetupRepository)

		groupNames := []string{"group1", "group2", "group3"}

		for _, group := range groupNames {
			meetupRepo.On("GetEventsUntilDateForGroup", ctx, group, now.AddDate(0, 6, 0)).
				Return(meetupFaker.CreateEvents(group, 2), nil)
			eventRepo.On("GetUpcomingEventsForGroup", ctx, group).
				Return(meetupFaker.CreateEvents(group, 1), nil)
			eventRepo.On("UpsertEvents", ctx, mock.Anything).Return(nil)
			eventRepo.On("ArchiveEvents", ctx, mock.Anything).Return(nil)
		}

		svc := NewService(
			ServiceConfig{GroupNames: groupNames},
			mockTimeSource,
			logging.NewMockLogger(),
			eventRepo,
			meetupRepo,
		)

		err := svc.Import(ctx)
		assert.NoError(t, err)

		meetupRepo.AssertNumberOfCalls(t, "GetEventsUntilDateForGroup", len(groupNames))
		eventRepo.AssertNumberOfCalls(t, "UpsertEvents", len(groupNames))
	})

	t.Run("archives missing events and upserts new ones", func(t *testing.T) {
		mockTimeSource := clock.NewMockTimeSource(now)
		eventRepo := new(MockEventRepository)
		meetupRepo := new(MockMeetupRepository)
		group := "test-group"

		savedEvents := meetupFaker.CreateEvents(group, 3)
		incomingEvents := savedEvents[1:]

		meetupRepo.On("GetEventsUntilDateForGroup", ctx, group, now.AddDate(0, 6, 0)).
			Return(incomingEvents, nil)
		eventRepo.On("GetUpcomingEventsForGroup", ctx, group).
			Return(savedEvents, nil)
		eventRepo.On("UpsertEvents", ctx, incomingEvents).Return(nil)
		eventRepo.On("ArchiveEvents", ctx, []string{savedEvents[0].ID}).Return(nil)

		svc := NewService(
			ServiceConfig{GroupNames: []string{group}},
			mockTimeSource,
			logging.NewMockLogger(),
			eventRepo,
			meetupRepo,
		)

		err := svc.Import(ctx)
		require.NoError(t, err)

		eventRepo.AssertExpectations(t)
	})

	t.Run("handles error fetching existing events", func(t *testing.T) {
		mockTimeSource := clock.NewMockTimeSource(now)
		eventRepo := new(MockEventRepository)
		meetupRepo := new(MockMeetupRepository)
		group := "error-group"
		cfg := ServiceConfig{GroupNames: []string{group}}
		expectedErr := errors.New("db error")

		eventRepo.On("GetUpcomingEventsForGroup", ctx, group).
			Return([]models.MeetupEvent{}, expectedErr)

		svc := NewService(
			cfg,
			mockTimeSource,
			logging.NewMockLogger(),
			eventRepo,
			meetupRepo,
		)

		err := svc.Import(ctx)
		assert.ErrorIs(t, err, expectedErr)

		meetupRepo.AssertNotCalled(t, "GetEventsUntilDateForGroup")
		eventRepo.AssertNotCalled(t, "UpsertEvents")
		eventRepo.AssertNotCalled(t, "ArchiveEvents")
	})

	t.Run("handles empty events scenario", func(t *testing.T) {
		mockTimeSource := clock.NewMockTimeSource(now)
		eventRepo := new(MockEventRepository)
		meetupRepo := new(MockMeetupRepository)
		group := "empty-group"
		cfg := ServiceConfig{GroupNames: []string{group}}

		meetupRepo.On("GetEventsUntilDateForGroup", ctx, group, now.AddDate(0, 6, 0)).
			Return([]models.MeetupEvent{}, nil)
		eventRepo.On("GetUpcomingEventsForGroup", ctx, group).
			Return([]models.MeetupEvent{}, nil)
		eventRepo.On("UpsertEvents", ctx, []models.MeetupEvent{}).Return(nil)
		eventRepo.On("ArchiveEvents", ctx, []string{}).Return(nil)

		svc := NewService(
			cfg,
			mockTimeSource,
			logging.NewMockLogger(),
			eventRepo,
			meetupRepo,
		)

		err := svc.Import(ctx)
		require.NoError(t, err)

		eventRepo.AssertCalled(t, "UpsertEvents", ctx, []models.MeetupEvent{})
		eventRepo.AssertNotCalled(t, "ArchiveEvents")
	})
}
