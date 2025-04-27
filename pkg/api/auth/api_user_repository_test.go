package auth

import (
	"github.com/stretchr/testify/assert"
	"sgf-meetup-api/pkg/api/apiconfig"
	"testing"
)

func TestNewDynamoDBAPIUserRepositoryConfig(t *testing.T) {
	cfg := &apiconfig.Config{
		APIUsersTableName: "apiUsers",
	}

	userRepoConfig := NewDynamoDBAPIUserRepositoryConfig(cfg)
	assert.Equal(t, cfg.APIUsersTableName, userRepoConfig.APIUserTable)
}
