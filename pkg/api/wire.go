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
)

var CommonSet = wire.NewSet(NewConfig, getLoggingConfig)
var DbSet = wire.NewSet(getDbConfig, db.NewClient)

func InitRouter(ctx context.Context) (*gin.Engine, error) {
	panic(wire.Build(
		//CommonSet,
		auth.ProviderSet,
		groupevents.ProviderSet,
		NewRouter,
	))
}
