//go:build wireinject
// +build wireinject

package meetupproxy

import (
	"github.com/google/wire"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/httpclient"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonSet = wire.NewSet(logging.DefaultLogger, clock.RealTimeSource, httpclient.DefaultClient, getLogLevel)

func InitService(config *Config) *Service {
	wire.Build(
		CommonSet,
		NewAuthHandler,
		NewServiceConfig,
		NewAuthHandlerConfig,
		NewService,
	)
	return &Service{}
}
