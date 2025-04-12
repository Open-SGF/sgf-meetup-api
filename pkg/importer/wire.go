//go:build wireinject
// +build wireinject

package importer

import (
	"context"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/clock"
	"sgf-meetup-api/pkg/db"
	"sgf-meetup-api/pkg/httpclient"
	"sgf-meetup-api/pkg/logging"
)

var CommonSet = wire.NewSet(logging.DefaultLogger, clock.RealTimeSource, httpclient.DefaultClient, getLogLevel)
var DBSet = wire.NewSet(getDbConfig, db.New)

func InitService(ctx context.Context, config *Config) (*Service, error) {
	wire.Build(
		CommonSet,
		DBSet,
		NewMeetupProxyGraphQLHandlerConfig,
		NewMeetupProxyGraphQLHandler,
		NewMeetupRepository,
		NewEventDBRepositoryConfig,
		NewEventDBRepository,
		NewServiceConfig,
		NewService,
	)
	return &Service{}, nil
}
