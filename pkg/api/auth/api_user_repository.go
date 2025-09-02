package auth

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/models"
)

type APIUserRepository interface {
	GetAPIUser(ctx context.Context, clientID string) (*models.APIUser, error)
}

type DynamoDBAPIUserRepositoryConfig struct {
	APIUserTable string
}

func NewDynamoDBAPIUserRepositoryConfig(config *apiconfig.Config) DynamoDBAPIUserRepositoryConfig {
	return DynamoDBAPIUserRepositoryConfig{
		APIUserTable: config.APIUsersTableName,
	}
}

type DynamoDBAPIUserRepository struct {
	config DynamoDBAPIUserRepositoryConfig
	db     *db.Client
}

func NewDynamoDBAPIUserRepository(
	config DynamoDBAPIUserRepositoryConfig,
	db *db.Client,
) *DynamoDBAPIUserRepository {
	return &DynamoDBAPIUserRepository{
		config: config,
		db:     db,
	}
}

func (r *DynamoDBAPIUserRepository) GetAPIUser(
	ctx context.Context,
	clientID string,
) (*models.APIUser, error) {
	result, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.config.APIUserTable),
		Key: map[string]types.AttributeValue{
			"clientId": &types.AttributeValueMemberS{Value: clientID},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, ErrAPIUserNotFound
	}

	var user models.APIUser
	if err = attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

var ErrAPIUserNotFound = errors.New("api user not found")

var APIUserRepositoryProviders = wire.NewSet(
	wire.Bind(new(APIUserRepository), new(*DynamoDBAPIUserRepository)),
	NewDynamoDBAPIUserRepositoryConfig,
	NewDynamoDBAPIUserRepository,
)
