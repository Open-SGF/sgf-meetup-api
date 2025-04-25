//go:build wireinject
// +build wireinject

package syncdynamodb

import (
	"context"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/syncdynamodb/syncdynamodbconfig"
)

var CommonProviders = wire.NewSet(syncdynamodbconfig.ConfigProviders, logging.DefaultLogger)

func InitService(ctx context.Context) (*Service, error) {
	panic(wire.Build(CommonProviders, db.Providers, NewService))
}
