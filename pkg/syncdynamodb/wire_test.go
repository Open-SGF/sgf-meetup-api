package syncdynamodb

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitService(t *testing.T) {
	ctx := context.Background()

	_, err := InitService(ctx)

	require.NoError(t, err)
}
