//go:build wireinject
// +build wireinject

package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/api/auth"
	"sgf-meetup-api/pkg/api/groupevents"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonSet = wire.NewSet(NewConfig, logging.DefaultLogger, getLoggingConfig)
var DbSet = wire.NewSet(getDbConfig, db.NewClient)

func InitRouter(ctx context.Context) (*gin.Engine, error) {
	panic(wire.Build(
		CommonSet,
		DbSet,
		auth.ProviderSet,
		groupevents.ProviderSet,
		NewRouter,
	))
}
