package auth

import (
	"testing"

	"sgf-meetup-api/pkg/api/apiconfig"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenValidatorConfig(t *testing.T) {
	cfg := &apiconfig.Config{
		JWTIssuer: "issuer",
		JWTSecret: []byte("secret"),
	}

	tokenConfig := NewTokenValidatorConfig(cfg)
	assert.Equal(t, cfg.JWTIssuer, tokenConfig.JWTIssuer)
	assert.Equal(t, cfg.JWTSecret, tokenConfig.JWTSecret)
}
