package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"sgf-meetup-api/pkg/db"
	"sgf-meetup-api/pkg/models"
	"time"
)

type Service struct {
	proxyFunctionName string
	eventsTable       string
	groupNames        []string
	dbOptions         db.Options
}

func New(proxyFunctionName, eventsTable string, groupNames []string, dbOptions db.Options) *Service {
	return &Service{
		proxyFunctionName: proxyFunctionName,
		eventsTable:       eventsTable,
		groupNames:        groupNames,
		dbOptions:         dbOptions,
	}
}

func NewFromConfig(c *Config) *Service {
	return New(c.MeetupProxyFunctionName, c.EventsTableName, c.MeetupGroupNames, db.Options{
		Endpoint:        c.DynamoDbEndpoint,
		Region:          c.AwsRegion,
		AccessKey:       c.AwsAccessKey,
		SecretAccessKey: c.AwsSecretAccessKey,
	})
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

func (s *Service) getEvents(ctx context.Context) ([]models.MeetupEvent, error) {
	fetchCutOff := time.Now().AddDate(0, 6, 0)
	var newestDate time.Time
}

func callMeetupGraphQL[T any](ctx context.Context, s *Service, query string, variables map[string]any) (*T, error) {
	request := struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}{
		Query:     query,
		Variables: variables,
	}

	requestBytes, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	responseBytes, err := callLambda(ctx, s.proxyFunctionName, requestBytes)

	if err != nil {
		return nil, err
	}

	var response T
	if err = json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil

}

func callLambda(ctx context.Context, functionName string, payload []byte) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	result, err := lambda.NewFromConfig(cfg).Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
	})

	if err != nil {
		return nil, err
	}

	if result.FunctionError != nil {
		return nil, fmt.Errorf("lambda execution error: %s", *result.FunctionError)
	}

	return result.Payload, nil
}
