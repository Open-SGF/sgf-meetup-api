package importer

import (
	"context"
	"errors"
	"log/slog"
	"sgf-meetup-api/pkg/shared/clock"
	"time"
)

type ServiceConfig struct {
	GroupNames []string
}

func NewServiceConfig(config *Config) ServiceConfig {
	return ServiceConfig{
		GroupNames: config.MeetupGroupNames,
	}
}

type Service struct {
	config           ServiceConfig
	timeSource       clock.TimeSource
	logger           *slog.Logger
	eventRepository  EventRepository
	meetupRepository MeetupRepository
}

func NewService(
	config ServiceConfig,
	timeSource clock.TimeSource,
	logger *slog.Logger,
	eventRepository EventRepository,
	meetupRepository MeetupRepository,
) *Service {
	return &Service{
		config:           config,
		timeSource:       timeSource,
		logger:           logger,
		eventRepository:  eventRepository,
		meetupRepository: meetupRepository,
	}
}

func (s *Service) Import(ctx context.Context) error {
	sixMonthsFromNow := s.timeSource.Now().AddDate(0, 6, 0)

	const maxConcurrency = 3
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(chan error, len(s.config.GroupNames))

	for _, group := range s.config.GroupNames {
		semaphore <- struct{}{}
		go s.importWorker(ctx, group, sixMonthsFromNow, results, semaphore)
	}

	var multiErr error
	for range s.config.GroupNames {
		if err := <-results; err != nil {
			multiErr = errors.Join(multiErr, err)
		}
	}
	close(results)
	return multiErr
}

func (s *Service) importWorker(ctx context.Context, group string, beforeDate time.Time, results chan<- error, signal <-chan struct{}) {
	defer func() { <-signal }()

	err := s.importForGroup(ctx, group, beforeDate)
	if err != nil {
		s.logger.Error("error fetching events", "group", group)
		results <- err
	} else {
		s.logger.Info("successfully imported events for group", "group", group)
		results <- nil
	}
}

func (s *Service) importForGroup(ctx context.Context, group string, beforeDate time.Time) error {
	savedEvents, err := s.eventRepository.GetUpcomingEventsForGroup(ctx, group)

	if err != nil {
		return err
	}

	missingEventIds := make([]string, 0)
	incomingEvents, err := s.meetupRepository.GetEventsUntilDateForGroup(ctx, group, beforeDate)

	if err != nil {
		return err
	}

	incomingEventIds := make(map[string]struct{}, len(incomingEvents))
	for _, incomingEvent := range incomingEvents {
		incomingEventIds[incomingEvent.ID] = struct{}{}
	}

	for _, savedEvent := range savedEvents {
		if _, ok := incomingEventIds[savedEvent.ID]; !ok {
			missingEventIds = append(missingEventIds, savedEvent.ID)
		}
	}

	if err = s.eventRepository.UpsertEvents(ctx, incomingEvents); err != nil {
		return err
	}

	if err = s.eventRepository.ArchiveEvents(ctx, missingEventIds); err != nil {
		return err
	}

	return nil
}
