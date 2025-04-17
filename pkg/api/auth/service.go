package auth

import (
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/models"
)

type ServiceConfig struct {
	//apiUsersTable string
}

type Service struct {
	//config ServiceConfig
	db *db.Client
}

func NewService(db *db.Client) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) AuthClientCredentials(clientId, clientSecret string) (models.AuthResult, error) {
	return models.AuthResult{}, nil
}
