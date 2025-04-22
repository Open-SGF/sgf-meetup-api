package groupevents

import (
	"context"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/shared/models"
	"time"
)

type ServiceConfig struct {
	AppURL string
}

func NewServiceConfig(config *apiconfig.Config) ServiceConfig {
	return ServiceConfig{
		AppURL: config.AppURL,
	}
}

type Service struct {
	config ServiceConfig
}

func NewService(config ServiceConfig) *Service {
	return &Service{
		config: config,
	}
}

type groupEventArgs struct {
	Before *time.Time
	After  *time.Time
	Cursor string
	Limit  *int
}

func (s *Service) GroupEvents(ctx context.Context, groupID string, args groupEventArgs) ([]models.MeetupEvent, *string, error) {
	return nil, nil, nil
}
