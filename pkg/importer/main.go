package importer

import (
	"context"
	"fmt"
	"log/slog"
	"sgf-meetup-api/pkg/clock"
	"sgf-meetup-api/pkg/db"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/logging"
	"time"
)

type Service struct {
	groupNames  []string
	timeSource  clock.TimeSource
	logger      *slog.Logger
	eventDBRepo EventDBRepository
	meetupRepo  MeetupRepository
}

func New(
	groupNames []string,
	timeSource clock.TimeSource,
	logger *slog.Logger,
	eventDBRepo EventDBRepository,
	meetupRepo MeetupRepository,
) *Service {
	return &Service{
		groupNames:  groupNames,
		timeSource:  timeSource,
		logger:      logger,
		eventDBRepo: eventDBRepo,
		meetupRepo:  meetupRepo,
	}
}

func NewFromConfig(c *Config) *Service {
	logger := logging.DefaultLogger(c.LogLevel)
	graphQLHandler := NewMeetupProxyGraphQLHandler(*infra.MeetupProxyFunctionName, logger)
	timeSource := clock.RealTimeSource()
	eventDBRepo := NewEventDBRepository(
		*infra.EventsTableProps.TableName,
		*infra.GroupUrlNameDateTimeIndex.IndexName,
		db.Options{
			Endpoint:        c.DynamoDbEndpoint,
			Region:          c.AwsRegion,
			AccessKey:       c.AwsAccessKey,
			SecretAccessKey: c.AwsSecretAccessKey,
		},
		timeSource,
		logger,
	)

	return New(
		c.MeetupGroupNames,
		timeSource,
		logger,
		eventDBRepo,
		NewMeetupRepository(graphQLHandler, logger),
	)
}

func (s *Service) Import(ctx context.Context) error {
	sixMonthsFromNow := s.timeSource.Now().AddDate(0, 6, 0)

	for _, group := range s.groupNames {
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
