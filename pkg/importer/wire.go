//go:build wireinject
// +build wireinject

package importer

import (
	"context"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/httpclient"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonSet = wire.NewSet(NewConfig, logging.DefaultLogger, clock.RealClockSet, httpclient.DefaultClient, getLoggingConfig)
var DBSet = wire.NewSet(getDbConfig, db.NewClient)
var EventRepositorySet = wire.NewSet(wire.Bind(new(EventRepository), new(*DynamoDBEventRepository)), NewDynamoDBEventRepositoryConfig, NewDynamoDBEventRepository)
var GraphQLHandlerSet = wire.NewSet(wire.Bind(new(GraphQLHandler), new(*LambdaProxyGraphQLHandler)), NewLambdaProxyGraphQLHandlerConfig, NewLambdaProxyGraphQLHandler)
var MeetupRepositorySet = wire.NewSet(wire.Bind(new(MeetupRepository), new(*GraphQLMeetupRepository)), NewGraphQLMeetupRepository)

func InitService(ctx context.Context) (*Service, error) {
	wire.Build(
		CommonSet,
		DBSet,
		EventRepositorySet,
		GraphQLHandlerSet,
		MeetupRepositorySet,
		NewServiceConfig,
		NewService,
	)
	return &Service{}, nil
}
