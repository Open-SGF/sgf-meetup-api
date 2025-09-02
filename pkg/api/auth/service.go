package auth

import (
	"context"
	"errors"
	"time"

	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/models"

	"golang.org/x/crypto/bcrypt"
)

type ServiceConfig struct {
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
}

func NewServiceConfig(config *apiconfig.Config) ServiceConfig {
	return ServiceConfig{
		AccessTokenExpiration:  time.Minute * 15,
		RefreshTokenExpiration: time.Hour * 24 * 30,
	}
}

type Service struct {
	config            ServiceConfig
	timeSource        clock.TimeSource
	apiUserRepository APIUserRepository
	tokenManager      TokenManager
}

func NewService(
	config ServiceConfig,
	timeSource clock.TimeSource,
	apiUserRepository APIUserRepository,
	tokenManager TokenManager,
) *Service {
	return &Service{
		config:            config,
		timeSource:        timeSource,
		apiUserRepository: apiUserRepository,
		tokenManager:      tokenManager,
	}
}

func (s *Service) AuthClientCredentials(
	ctx context.Context,
	clientID, clientSecret string,
) (*models.AuthResult, error) {
	user, err := s.apiUserRepository.GetAPIUser(ctx, clientID)

	if errors.Is(err, ErrAPIUserNotFound) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	if !s.verifyClientSecret(clientSecret, user.HashedClientSecret) {
		return nil, ErrInvalidCredentials
	}

	return s.getAuthResult(clientID)
}

func (s *Service) RefreshCredentials(
	ctx context.Context,
	refreshToken string,
) (*models.AuthResult, error) {
	token, err := s.tokenManager.Validate(refreshToken)
	if err != nil {
		return nil, err
	}

	_, err = s.apiUserRepository.GetAPIUser(ctx, token.ClientID)

	if errors.Is(err, ErrAPIUserNotFound) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	return s.getAuthResult(token.ClientID)
}

func (s *Service) getAuthResult(clientID string) (*models.AuthResult, error) {
	now := s.timeSource.Now()
	accessTokenExpiresAt := now.Add(s.config.AccessTokenExpiration)
	refreshTokenExpiresAt := now.Add(s.config.RefreshTokenExpiration)

	accessToken, err := s.tokenManager.CreateSignedToken(clientID, accessTokenExpiresAt)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenManager.CreateSignedToken(clientID, refreshTokenExpiresAt)
	if err != nil {
		return nil, err
	}

	return &models.AuthResult{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}, nil
}

func (s *Service) verifyClientSecret(clientSecret string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(clientSecret))
	return err == nil
}

var ErrInvalidCredentials = errors.New("provided credentials are invalid")
