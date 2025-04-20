package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/models"
	"time"
)

type ServiceConfig struct {
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
	jwtIssuer              string
	jwtSecret              []byte
	apiUsersTable          string
}

func NewServiceConfig(config *apiconfig.Config) ServiceConfig {
	return ServiceConfig{
		accessTokenExpiration:  time.Minute * 15,
		refreshTokenExpiration: time.Hour * 24 * 30,
		jwtIssuer:              config.JWTIssuer,
		jwtSecret:              []byte(config.JWTSecret),
		apiUsersTable:          config.ApiUsersTableName,
	}
}

type Service struct {
	config     ServiceConfig
	db         *db.Client
	timeSource clock.TimeSource
}

func NewService(config ServiceConfig, db *db.Client, timeSource clock.TimeSource) *Service {
	return &Service{
		config:     config,
		db:         db,
		timeSource: timeSource,
	}
}

func (s *Service) AuthClientCredentials(ctx context.Context, clientID, clientSecret string) (*models.AuthResult, error) {
	user, err := s.getApiUser(ctx, clientID)

	if errors.Is(err, ApiUserNotFound) {
		return nil, InvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	if !s.verifyClientSecret(clientSecret, user.HashedClientSecret) {
		return nil, InvalidCredentials
	}

	return s.getAuthResult(clientID)
}

func (s *Service) RefreshCredentials(ctx context.Context, refreshToken string) (*models.AuthResult, error) {
	clientID, err := s.validateToken(refreshToken)

	if err != nil {
		return nil, err
	}

	_, err = s.getApiUser(ctx, clientID)

	if errors.Is(err, ApiUserNotFound) {
		return nil, InvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	return s.getAuthResult(clientID)
}

func (s *Service) getAuthResult(clientID string) (*models.AuthResult, error) {
	now := s.timeSource.Now()
	accessTokenExpiresAt := now.Add(s.config.accessTokenExpiration)
	refreshTokenExpiresAt := now.Add(s.config.refreshTokenExpiration)

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

func (s *Service) getApiUser(ctx context.Context, clientID string) (*models.ApiUser, error) {
	result, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.config.apiUsersTable),
		Key: map[string]types.AttributeValue{
			"clientId": &types.AttributeValueMemberS{Value: clientID},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, ApiUserNotFound
	}

	var user models.ApiUser
	if err = attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) verifyClientSecret(clientSecret, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(clientSecret))
	return err == nil
}

func (s *Service) createToken(clientID string, expiration time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    s.config.jwtIssuer,
		Subject:   clientID,
		IssuedAt:  jwt.NewNumericDate(s.timeSource.Now()),
		ExpiresAt: jwt.NewNumericDate(expiration),
	}

	signedJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return signedJWT.SignedString(s.config.jwtSecret)
}

func (s *Service) validateToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing met hod: %v", token.Header["alg"])
		}
		return s.config.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", InvalidCredentials
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", InvalidCredentials
	}

	clientID := claims.Audience[0]

	if clientID == "" {
		return "", InvalidCredentials
	}

	return clientID, nil
}

var InvalidCredentials = errors.New("provided credentials are invalid")
var ApiUserNotFound = errors.New("api user not found")
