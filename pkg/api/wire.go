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
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonSet = wire.NewSet(apiconfig.NewConfig, logging.DefaultLogger, apiconfig.NewLoggingConfig)
var DbSet = wire.NewSet(apiconfig.NewDBConfig, db.NewClient)

func InitRouter(ctx context.Context) (*gin.Engine, error) {
	panic(wire.Build(
		CommonSet,
		DbSet,
		auth.ProviderSet,
		groupevents.ProviderSet,
		NewRouter,
	))
}
