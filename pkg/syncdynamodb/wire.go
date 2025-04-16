//go:build wireinject
// +build wireinject

package syncdynamodb

import (
	"context"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
)

var CommonSet = wire.NewSet(NewConfig, logging.DefaultLogger, getLoggingConfig)
var DBSet = wire.NewSet(getDbConfig, db.NewClient)

func InitService(ctx context.Context) (*Service, error) {
	wire.Build(CommonSet, DBSet, NewService)
	return &Service{}, nil
}
