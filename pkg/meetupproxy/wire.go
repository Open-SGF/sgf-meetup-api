//go:build wireinject
// +build wireinject

package meetupproxy

import (
	"github.com/google/wire"
	"sgf-meetup-api/pkg/clock"
	"sgf-meetup-api/pkg/httpclient"
	"sgf-meetup-api/pkg/logging"
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
