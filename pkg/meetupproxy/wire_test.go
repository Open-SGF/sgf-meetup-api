package meetupproxy

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitService(t *testing.T) {
	t.Setenv("MEETUP_PRIVATE_KEY_BASE64", "c29tZUJhc2U2NEtleQ==")
	t.Setenv("MEETUP_USER_ID", "meetupUserId")
	t.Setenv("MEETUP_CLIENT_KEY", "meetupClientKey")
	t.Setenv("MEETUP_SIGNING_KEY_ID", "signingKeyId")

	_, err := InitService()

	require.NoError(t, err)
}
