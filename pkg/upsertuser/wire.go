//go:build wireinject
// +build wireinject

package upsertuser

import (
	"context"

	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/upsertuser/upsertuserconfig"

	"github.com/google/wire"
)

func InitService(ctx context.Context) (*Service, error) {
	panic(wire.Build(
		logging.DefaultLogger,
		upsertuserconfig.ConfigProviders,
		db.Providers,
		NewService,
	))
}
