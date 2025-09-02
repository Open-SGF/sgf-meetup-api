package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sgf-meetup-api/pkg/api/apiconfig"
)

func TestNewServiceConfig(t *testing.T) {
	cfg := &apiconfig.Config{}

	serviceConfig := NewServiceConfig(cfg)

	assert.Greater(t, int(serviceConfig.AccessTokenExpiration), 0)
	assert.Greater(t, int(serviceConfig.RefreshTokenExpiration), 0)
}
