package importer

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
	"sgf-meetup-api/pkg/clock"
	"sgf-meetup-api/pkg/db"
	"time"
)

type Service struct {
	eventsTable      string
	groupNames       []string
	dbOptions        db.Options
	timeSource       clock.TimeSource
	meetupRepository MeetupRepository
}

func New(eventsTable string, groupNames []string, dbOptions db.Options, timeSource clock.TimeSource, meetupRepository MeetupRepository) *Service {
	return &Service{
		eventsTable:      eventsTable,
		groupNames:       groupNames,
		dbOptions:        dbOptions,
		timeSource:       timeSource,
		meetupRepository: meetupRepository,
	}
}

func NewFromConfig(c *Config) *Service {
	return New(
		c.EventsTableName,
		c.MeetupGroupNames,
		db.Options{
			Endpoint:        c.DynamoDbEndpoint,
			Region:          c.AwsRegion,
			AccessKey:       c.AwsAccessKey,
			SecretAccessKey: c.AwsSecretAccessKey,
		},
		clock.RealTimeSource(),
		NewMeetupRepository(NewMeetupProxyGraphQLHandler(c.MeetupProxyFunctionName)),
	)
}

func (s *Service) Import(ctx context.Context) error {
	sixMonthsFromNow := s.timeSource.Now().AddDate(0, 6, 0)

	for _, group := range s.groupNames {
		err := s.importForGroup(ctx, group, sixMonthsFromNow)

		if err != nil {
			log.Println("")
		}
	}

	db, err := db.New(ctx, &s.dbOptions)

	if err != nil {
		return err
	}

	result, err := db.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(s.eventsTable),
	})

	if err != nil {
		return err
	}

	fmt.Println(result.Items)

	return nil
}

func (s *Service) importForGroup(ctx context.Context, group string, beforeDate time.Time) error {
	//events, err := s.meetupRepository.GetEventsUntilDateForGroup(ctx, group, beforeDate)
	return nil
}
