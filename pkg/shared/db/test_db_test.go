package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/infra"
	"testing"
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

		assert.Equal(t, 1, getRecordCount(t, ctx, testDB.Client, *infra.ApiUsersTableProps.TableName))

		err = testDB.Reset(ctx)

		require.NoError(t, err)

		assert.Equal(t, 0, getRecordCount(t, ctx, testDB.Client, *infra.ApiUsersTableProps.TableName))
	})
}

func getRecordCount(t *testing.T, ctx context.Context, client *Client, tableName string) int {
	var totalCount = 0
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:         aws.String(tableName),
			Select:            types.SelectCount,
			ExclusiveStartKey: lastEvaluatedKey,
		}

		result, err := client.Scan(ctx, input)
		require.NoError(t, err)

		totalCount += int(result.Count)

		if result.LastEvaluatedKey == nil {
			break
		}

		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return totalCount
}
