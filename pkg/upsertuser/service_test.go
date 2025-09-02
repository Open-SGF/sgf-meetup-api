package upsertuser

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/upsertuser/upsertuserconfig"
)

func TestService_UpsertUser(t *testing.T) {
	ctx := context.Background()

	testDB, err := db.NewTestDB(ctx)
	require.NoError(t, err)

	service := NewService(&upsertuserconfig.Config{}, testDB.Client)

	tableName := *infra.ApiUsersTableProps.TableName

	t.Run("validates client secret", func(t *testing.T) {
		tests := []struct {
			name   string
			secret string
		}{
			{
				name:   fmt.Sprintf("must be at least %d characters", minLength),
				secret: "lowchars",
			},
			{
				name:   "contain at least one uppercase letter",
				secret: "onlylowercasechars",
			},
			{
				name:   "contain at least one lowercase letter",
				secret: "ONLYUPPERCASECHARS",
			},
			{
				name:   "contain at least one number",
				secret: "noNUMBERSbutLOTSofCHARS",
			},
			{
				name:   "contain at least one special character",
				secret: "UPPERCASElowercase1234",
			},
			{
				name:   "contains invalid characters",
				secret: "UPPERCASElowercase1234!!`",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := service.UpsertUser(ctx, tableName, "client", test.secret)

				assert.Contains(t, err.Error(), test.name)
			})
		}
	})

	t.Run("creates new user", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		clientID := "test-user"
		err := service.UpsertUser(ctx, tableName, clientID, "UPPERCASElowercase1234!!")
		require.NoError(t, err)

		testDB.CheckItemExists(ctx, tableName, "clientId", clientID)
	})

	t.Run("creates new user", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		clientID := "test-user"
		err := service.UpsertUser(ctx, tableName, clientID, "UPPERCASElowercase1234!!")
		require.NoError(t, err)

		err = service.UpsertUser(ctx, tableName, clientID, "UPPERCASElowercase1!")
		require.NoError(t, err)

		require.Equal(t, 1, testDB.GetItemCount(ctx, tableName))
	})
}
