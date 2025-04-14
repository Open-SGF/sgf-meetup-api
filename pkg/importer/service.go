package importer

import (
	"context"
	"log/slog"
	"sgf-meetup-api/pkg/shared/clock"
	"sync"
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
	config          ServiceConfig
	timeSource      clock.TimeSource
	logger          *slog.Logger
	eventRepository EventRepository
	meetupRepo      MeetupRepository
}

func NewService(
	config ServiceConfig,
	timeSource clock.TimeSource,
	logger *slog.Logger,
	eventRepository EventRepository,
	meetupRepo MeetupRepository,
) *Service {
	return &Service{
		config:          config,
		timeSource:      timeSource,
		logger:          logger,
		eventRepository: eventRepository,
		meetupRepo:      meetupRepo,
	}
}

func (s *Service) Import(ctx context.Context) error {
	sixMonthsFromNow := s.timeSource.Now().AddDate(0, 6, 0)

	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup

	for _, group := range s.config.GroupNames {
		sem <- struct{}{}
		wg.Add(1)

		go func(g string) {
			defer func() {
				<-sem
				wg.Done()
			}()

			err := s.importForGroup(ctx, g, sixMonthsFromNow)
			if err != nil {
				s.logger.Error("error fetching events", "group", g)
			} else {
				s.logger.Info("successfully imported events for group", "group", g)
			}
		}(group)
	}

	wg.Wait()
	return nil
}

func (s *Service) importForGroup(ctx context.Context, group string, beforeDate time.Time) error {
	savedEvents, err := s.eventRepository.GetUpcomingEventsForGroup(ctx, group)

	if err != nil {
		return err
	}

	missingEventIds := make([]string, 0)
	incomingEvents, err := s.meetupRepo.GetEventsUntilDateForGroup(ctx, group, beforeDate)

	if err != nil {
		return err
	}

	incomingEventIds := make(map[string]bool, len(incomingEvents))
	for _, incomingEvent := range incomingEvents {
		incomingEventIds[incomingEvent.ID] = true
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
