package importer

import (
	"context"
	"fmt"
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
	config      ServiceConfig
	timeSource  clock.TimeSource
	logger      *slog.Logger
	eventDBRepo EventRepository
	meetupRepo  MeetupRepository
}

func NewService(
	config ServiceConfig,
	timeSource clock.TimeSource,
	logger *slog.Logger,
	eventDBRepo EventRepository,
	meetupRepo MeetupRepository,
) *Service {
	return &Service{
		config:      config,
		timeSource:  timeSource,
		logger:      logger,
		eventDBRepo: eventDBRepo,
		meetupRepo:  meetupRepo,
	}
}

func (s *Service) Import(ctx context.Context) error {
	sixMonthsFromNow := s.timeSource.Now().AddDate(0, 6, 0)

	for _, group := range s.config.GroupNames {
		err := s.importForGroup(ctx, group, sixMonthsFromNow)

		if err != nil {
			s.logger.Error("error fetching events for group", "group", group)
		} else {
			s.logger.Info("successfully imported events for group", "group", group)
		}
	}

	return nil
}

func (s *Service) importForGroup(ctx context.Context, group string, beforeDate time.Time) error {
	savedEvents, err := s.eventDBRepo.GetUpcomingEventsForGroup(ctx, group)

	if err != nil {
		return err
	}

	fmt.Println(savedEvents)

	//events, err := s.meetupRepo.GetEventsUntilDateForGroup(ctx, group, beforeDate)
	return nil
}
