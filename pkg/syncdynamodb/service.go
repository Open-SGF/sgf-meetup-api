package syncdynamodb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log/slog"
	"sgf-meetup-api/pkg/shared/db"
)

type Service struct {
	db     *dynamodb.Client
	logger *slog.Logger
}

func NewService(db *dynamodb.Client, logger *slog.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

func (s *Service) Run(ctx context.Context) error {
	return db.SyncTables(ctx, s.db, s.logger)
}
