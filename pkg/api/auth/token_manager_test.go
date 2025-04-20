package auth

import (
	"github.com/stretchr/testify/assert"
	"sgf-meetup-api/pkg/api/apiconfig"
	"testing"
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
