package auth

import (
	"github.com/stretchr/testify/assert"
	"sgf-meetup-api/pkg/api/apiconfig"
	"testing"
)

func TestNewServiceConfig(t *testing.T) {
	cfg := &apiconfig.Config{
		JWTIssuer: "issuer",
		JWTSecret: "secret",
	}

	serviceConfig := NewServiceConfig(cfg)

	assert.Equal(t, cfg.JWTIssuer, serviceConfig.JWTIssuer)
	assert.Equal(t, cfg.JWTSecret, string(serviceConfig.JWTSecret))
	assert.Greater(t, int(serviceConfig.AccessTokenExpiration), 0)
	assert.Greater(t, int(serviceConfig.RefreshTokenExpiration), 0)
}
