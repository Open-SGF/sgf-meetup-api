package syncdynamodb

import (
	"context"
	"log/slog"
	"sgf-meetup-api/pkg/infra/customconstructs"
	"sgf-meetup-api/pkg/shared/db"
)

type Service struct {
	db     *db.Client
	logger *slog.Logger
}

func NewService(db *db.Client, logger *slog.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

func (s *Service) Run(ctx context.Context, tables []customconstructs.DynamoTableProps) error {
	return db.SyncTables(ctx, s.logger, s.db, tables)
}
