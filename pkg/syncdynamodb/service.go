package syncdynamodb

import (
	"context"
	"log/slog"
	"sgf-meetup-api/pkg/infra/customconstructs"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/syncdynamodb/syncdynamodbconfig"
)

type Service struct {
	db     *db.Client
	logger *slog.Logger
	config *syncdynamodbconfig.Config
}

func NewService(config *syncdynamodbconfig.Config, db *db.Client, logger *slog.Logger) *Service {
	return &Service{
		config: config,
		db:     db,
		logger: logger,
	}
}

func (s *Service) Run(ctx context.Context, tables []customconstructs.DynamoTableProps) error {
	return db.SyncTables(ctx, s.logger, s.db, s.config.AppEnv, tables)
}
