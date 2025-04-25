//go:build wireinject
// +build wireinject

package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/api/auth"
	"sgf-meetup-api/pkg/api/groupevents"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonProviders = wire.NewSet(apiconfig.ConfigProviders, logging.DefaultLogger, clock.RealClockProvider)
var DBProvider = wire.NewSet(db.NewClient)

func InitRouter(ctx context.Context) (*gin.Engine, error) {
	panic(wire.Build(
		CommonProviders,
		DBProvider,
		auth.Providers,
		groupevents.Providers,
		NewRouter,
	))
}
