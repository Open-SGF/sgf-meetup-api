//go:build wireinject
// +build wireinject

package upsertuser

import (
	"context"

	"github.com/google/wire"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/upsertuser/upsertuserconfig"
)

func InitService(ctx context.Context) (*Service, error) {
	panic(wire.Build(
		logging.DefaultLogger,
		upsertuserconfig.ConfigProviders,
		db.Providers,
		NewService,
	))
}
