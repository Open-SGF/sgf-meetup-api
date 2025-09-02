package auth

import (
	"testing"

	"sgf-meetup-api/pkg/api/apiconfig"

	"github.com/stretchr/testify/assert"
)

func TestNewServiceConfig(t *testing.T) {
	cfg := &apiconfig.Config{}

	serviceConfig := NewServiceConfig(cfg)

	assert.Greater(t, int(serviceConfig.AccessTokenExpiration), 0)
	assert.Greater(t, int(serviceConfig.RefreshTokenExpiration), 0)
}
