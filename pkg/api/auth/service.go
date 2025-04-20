package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/models"
	"time"
)

type ServiceConfig struct {
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
	JWTIssuer              string
	JWTSecret              []byte
}

func NewServiceConfig(config *apiconfig.Config) ServiceConfig {
	return ServiceConfig{
		AccessTokenExpiration:  time.Minute * 15,
		RefreshTokenExpiration: time.Hour * 24 * 30,
		JWTIssuer:              config.JWTIssuer,
		JWTSecret:              []byte(config.JWTSecret),
	}
}

type Service struct {
	config            ServiceConfig
	timeSource        clock.TimeSource
	apiUserRepository APIUserRepository
}

func NewService(config ServiceConfig, timeSource clock.TimeSource, apiUserRepository APIUserRepository) *Service {
	return &Service{
		config:            config,
		timeSource:        timeSource,
		apiUserRepository: apiUserRepository,
	}
}

func (s *Service) AuthClientCredentials(ctx context.Context, clientID, clientSecret string) (*models.AuthResult, error) {
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

func (s *Service) RefreshCredentials(ctx context.Context, refreshToken string) (*models.AuthResult, error) {
	clientID, err := s.validateToken(refreshToken)

	if err != nil {
		return nil, err
	}

	_, err = s.apiUserRepository.GetAPIUser(ctx, clientID)

	if errors.Is(err, ErrAPIUserNotFound) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	return s.getAuthResult(clientID)
}

func (s *Service) getAuthResult(clientID string) (*models.AuthResult, error) {
	now := s.timeSource.Now()
	accessTokenExpiresAt := now.Add(s.config.AccessTokenExpiration)
	refreshTokenExpiresAt := now.Add(s.config.RefreshTokenExpiration)

	accessToken, err := s.createToken(clientID, accessTokenExpiresAt)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.createToken(clientID, refreshTokenExpiresAt)
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

func (s *Service) verifyClientSecret(clientSecret, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(clientSecret))
	return err == nil
}

func (s *Service) createToken(clientID string, expiration time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    s.config.JWTIssuer,
		Subject:   clientID,
		IssuedAt:  jwt.NewNumericDate(s.timeSource.Now()),
		ExpiresAt: jwt.NewNumericDate(expiration),
	}

	signedJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return signedJWT.SignedString(s.config.JWTSecret)
}

func (s *Service) validateToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing met hod: %v", token.Header["alg"])
		}
		return s.config.JWTSecret, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrInvalidCredentials
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", ErrInvalidCredentials
	}

	clientID := claims.Audience[0]

	if clientID == "" {
		return "", ErrInvalidCredentials
	}

	return clientID, nil
}

var ErrInvalidCredentials = errors.New("provided credentials are invalid")
