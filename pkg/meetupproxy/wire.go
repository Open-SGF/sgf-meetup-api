//go:build wireinject
// +build wireinject

package meetupproxy

import (
	"context"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/meetupproxy/meetupproxyconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/httpclient"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonProviders = wire.NewSet(meetupproxyconfig.ConfigProviders, logging.DefaultLogger, clock.RealClockProvider, httpclient.DefaultClient)

func InitService(ctx context.Context) (*Service, error) {
	panic(wire.Build(
		CommonProviders,
		AuthHandlerProviders,
		NewServiceConfig,
		NewService,
	))
}
