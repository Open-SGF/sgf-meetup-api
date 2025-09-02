//go:build wireinject
// +build wireinject

package importer

import (
	"context"

	"github.com/google/wire"
	"sgf-meetup-api/pkg/importer/importerconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/httpclient"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonProviders = wire.NewSet(
	importerconfig.ConfigProviders,
	logging.DefaultLogger,
	clock.RealClockProvider,
	httpclient.DefaultClient,
	db.Providers,
)

func InitService(ctx context.Context) (*Service, error) {
	panic(wire.Build(
		CommonProviders,
		EventRepositoryProviders,
		GraphQLHandlerProviders,
		MeetupRepositoryProviders,
		NewServiceConfig,
		NewService,
	))
}
