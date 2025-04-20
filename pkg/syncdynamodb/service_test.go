package syncdynamodb

import (
	"context"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"
)

func TestService_Run(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.NewTestDBWithoutMigrations(ctx)

	require.NoError(t, err)
	defer testDB.Close()

	logger := logging.NewMockLogger()

	service := NewService(testDB.Client, logger)

	err = service.Run(ctx, infra.Tables)

	require.NoError(t, err)
}
