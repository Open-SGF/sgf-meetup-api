//go:build wireinject
// +build wireinject

package meetupproxy

import (
	"github.com/google/wire"
	"sgf-meetup-api/pkg/meetupproxy/meetupproxyconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/httpclient"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonSet = wire.NewSet(meetupproxyconfig.NewConfig, logging.DefaultLogger, clock.RealClockProvider, httpclient.DefaultClient, meetupproxyconfig.NewLoggingConfig)
var AuthHandlerSet = wire.NewSet(wire.Bind(new(AuthHandler), new(*MeetupHttpAuthHandler)), NewMeetupAuthHandlerConfig, NewMeetupHttpAuthHandler)

func InitService() (*Service, error) {
	panic(wire.Build(
		CommonSet,
		AuthHandlerSet,
		NewServiceConfig,
		NewService,
	))
}
