package db

import (
	"context"
	"testing"

	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/models"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestDB(t *testing.T) {
	ctx := context.Background()
	testDB, err := NewTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Close()

	t.Run("creates the correct number of tables", func(t *testing.T) {
		count := 0
		var lastEvaluatedTableName *string

		for {
			input := &dynamodb.ListTablesInput{
				ExclusiveStartTableName: lastEvaluatedTableName,
			}

			result, err := testDB.Client.ListTables(ctx, input)
			require.NoError(t, err)

			count += len(result.TableNames)

			if result.LastEvaluatedTableName == nil {
				break
			}
			lastEvaluatedTableName = result.LastEvaluatedTableName
		}

		assert.Equal(t, len(infra.Tables), count)
	})

	t.Run("resets data in tables", func(t *testing.T) {
		_, err := testDB.Client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: infra.ApiUsersTableProps.TableName,
			Item: map[string]types.AttributeValue{
				"clientId": &types.AttributeValueMemberS{Value: "someId"},
			},
		})

		require.NoError(t, err)

		assert.Equal(t, 1, testDB.GetItemCount(ctx, *infra.ApiUsersTableProps.TableName))

		err = testDB.Reset(ctx)

		require.NoError(t, err)

		assert.Equal(t, 0, testDB.GetItemCount(ctx, *infra.ApiUsersTableProps.TableName))
	})

	t.Run("inserts test items", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		items := []models.APIUser{
			{ClientID: "1", HashedClientSecret: []byte("test")},
			{ClientID: "3", HashedClientSecret: []byte("test")},
		}

		testDB.InsertTestItems(ctx, *infra.ApiUsersTableProps.TableName, items)

		assert.Equal(t, len(items), testDB.GetItemCount(ctx, *infra.ApiUsersTableProps.TableName))
	})
}
