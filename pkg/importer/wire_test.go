package importer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitService(t *testing.T) {
	ctx := context.Background()

	t.Setenv("MEETUP_PROXY_FUNCTION_NAME", "meetupproxy")
	t.Setenv("ARCHIVED_EVENTS_TABLE_NAME", "archived-events")
	t.Setenv("EVENTS_TABLE_NAME", "events")
	t.Setenv("GROUP_ID_DATE_TIME_INDEX_NAME", "group-index")

	_, err := InitService(ctx)

	require.NoError(t, err)
}
