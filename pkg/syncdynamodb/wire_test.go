package syncdynamodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitService(t *testing.T) {
	ctx := context.Background()

	_, err := InitService(ctx)

	require.NoError(t, err)
}
