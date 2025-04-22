package groupevents

import (
	"github.com/stretchr/testify/assert"
	"sgf-meetup-api/pkg/api/apiconfig"
	"testing"
)

func TestNewServiceConfig(t *testing.T) {
	cfg := &apiconfig.Config{
		AppURL: "http://localhost",
	}

	serviceConfig := NewServiceConfig(cfg)

	assert.Equal(t, cfg.AppURL, serviceConfig.AppURL)
}
