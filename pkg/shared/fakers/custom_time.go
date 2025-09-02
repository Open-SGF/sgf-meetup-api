package fakers

import (
	"sgf-meetup-api/pkg/shared/models"

	"github.com/brianvoe/gofakeit/v7"
)

func init() {
	gofakeit.AddFuncLookup("future_customtime", gofakeit.Info{
		Category:    "custom",
		Description: "A CustomTime instance in the future",
		Output:      "CustomTime",
		Generate: func(f *gofakeit.Faker, m *gofakeit.MapParams, info *gofakeit.Info) (any, error) {
			t := f.FutureDate()
			return models.CustomTime{Time: t}, nil
		},
	})
}
