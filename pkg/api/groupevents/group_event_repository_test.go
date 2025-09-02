package groupevents

import (
	"testing"

	"sgf-meetup-api/pkg/api/apiconfig"

	"github.com/stretchr/testify/assert"
)

func TestNewDynamoDBGroupEventRepositoryConfig(t *testing.T) {
	cfg := &apiconfig.Config{
		EventsTableName:          "events",
		GroupIDDateTimeIndexName: "groupIndex",
	}

	repoConfig := NewDynamoDBGroupEventRepositoryConfig(cfg)

	assert.Equal(t, cfg.EventsTableName, repoConfig.EventsTableName)
	assert.Equal(t, cfg.GroupIDDateTimeIndexName, repoConfig.GroupDateIndexName)
}
