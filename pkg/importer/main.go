package importer

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"sgf-meetup-api/pkg/db"
)

type Service struct {
	eventsTable      string
	groupNames       []string
	dbOptions        db.Options
	meetupRepository MeetupRepository
}

func New(eventsTable string, groupNames []string, dbOptions db.Options, meetupRepository MeetupRepository) *Service {
	return &Service{
		eventsTable:      eventsTable,
		groupNames:       groupNames,
		dbOptions:        dbOptions,
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
		NewMeetupRepository(NewMeetupProxyGraphQLHandler(c.MeetupProxyFunctionName)),
	)
}

func (s *Service) Import(ctx context.Context) error {
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
