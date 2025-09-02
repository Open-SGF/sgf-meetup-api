package syncdynamodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/syncdynamodb/syncdynamodbconfig"
)

func TestService_Run(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.NewTestDBWithoutMigrations(ctx)

	require.NoError(t, err)
	defer testDB.Close()

	logger := logging.NewMockLogger()

	config := &syncdynamodbconfig.Config{}
	service := NewService(config, testDB.Client, logger)

	err = service.Run(ctx, infra.Tables)

	require.NoError(t, err)
}
