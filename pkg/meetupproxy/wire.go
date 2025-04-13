//go:build wireinject
// +build wireinject

package meetupproxy

import (
	"github.com/google/wire"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/httpclient"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonSet = wire.NewSet(logging.DefaultLogger, clock.RealClockSet, httpclient.DefaultClient, getLoggingConfig)
var AuthHandlerSet = wire.NewSet(wire.Bind(new(AuthHandler), new(*MeetupHttpAuthHandler)), NewMeetupAuthHandlerConfig, NewMeetupHttpAuthHandler)

func InitService(config *Config) *Service {
	wire.Build(
		CommonSet,
		AuthHandlerSet,
		NewServiceConfig,
		NewService,
	)
	return &Service{}
}
