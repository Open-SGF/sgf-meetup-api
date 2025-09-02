package api

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestInitRouter(t *testing.T) {
	ctx := context.Background()

	gin.SetMode(gin.TestMode)

	t.Setenv("EVENTS_TABLE_NAME", "events")
	t.Setenv("API_USERS_TABLE_NAME", "users")
	t.Setenv("GROUP_ID_DATE_TIME_INDEX_NAME", "group-index")
	t.Setenv("JWT_SECRET", "secretkey")

	_, err := InitRouter(ctx)

	require.NoError(t, err)
}
