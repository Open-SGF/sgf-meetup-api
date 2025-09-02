package auth

import (
	"testing"

	"sgf-meetup-api/pkg/api/apiconfig"

	"github.com/stretchr/testify/assert"
)

func TestNewDynamoDBAPIUserRepositoryConfig(t *testing.T) {
	cfg := &apiconfig.Config{
		APIUsersTableName: "apiUsers",
	}

	userRepoConfig := NewDynamoDBAPIUserRepositoryConfig(cfg)
	assert.Equal(t, cfg.APIUsersTableName, userRepoConfig.APIUserTable)
}
