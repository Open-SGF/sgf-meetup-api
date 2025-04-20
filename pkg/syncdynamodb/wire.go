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

var CommonSet = wire.NewSet(syncdynamodbconfig.NewConfig, logging.DefaultLogger, syncdynamodbconfig.NewLoggingConfig)
var DBSet = wire.NewSet(syncdynamodbconfig.NewDBConfig, db.NewClient)

func InitService(ctx context.Context) (*Service, error) {
	panic(wire.Build(CommonSet, DBSet, NewService))
}
