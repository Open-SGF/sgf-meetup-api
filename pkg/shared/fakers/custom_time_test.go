package fakers

import (
	"testing"
	"time"

	"sgf-meetup-api/pkg/shared/models"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFutureCustomTime(t *testing.T) {
	type someStruct struct {
		Time *models.CustomTime `fake:"{future_customtime}"`
	}

	now := time.Now()
	faker := gofakeit.New(0)

	var instance someStruct
	err := faker.Struct(&instance)
	require.NoError(t, err)

	assert.Greater(t, instance.Time.Time, now)
}
