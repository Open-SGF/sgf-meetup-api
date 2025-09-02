package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/shared/clock"
)

type TokenManagerConfig struct {
	JWTIssuer string
	JWTSecret []byte
}

func NewTokenValidatorConfig(config *apiconfig.Config) TokenManagerConfig {
	return TokenManagerConfig{
		JWTIssuer: config.JWTIssuer,
		JWTSecret: config.JWTSecret,
	}
}

type TokenManager interface {
	Validate(tokenStr string) (*ParsedToken, error)
	CreateSignedToken(subject string, expiration time.Time) (string, error)
}

type TokenManagerImpl struct {
	config     TokenManagerConfig
	timeSource clock.TimeSource
}

func NewTokenManager(config TokenManagerConfig, timeSource clock.TimeSource) *TokenManagerImpl {
	return &TokenManagerImpl{
		config:     config,
		timeSource: timeSource,
	}
}

type ParsedToken struct {
	ClientID string
}

func (tm *TokenManagerImpl) Validate(tokenStr string) (*ParsedToken, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return tm.config.JWTSecret, nil
		},
		jwt.WithTimeFunc(tm.timeSource.Now),
	)

	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, ErrInvalidCredentials
	}

	if claims.Subject == "" {
		return nil, ErrInvalidCredentials
	}

	return &ParsedToken{
		ClientID: claims.Subject,
	}, nil
}

func (tm *TokenManagerImpl) CreateSignedToken(
	clientID string,
	expiration time.Time,
) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    tm.config.JWTIssuer,
		Subject:   clientID,
		IssuedAt:  jwt.NewNumericDate(tm.timeSource.Now()),
		ExpiresAt: jwt.NewNumericDate(expiration),
	}

	signedJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return signedJWT.SignedString(tm.config.JWTSecret)
}

var TokenValidatorProviders = wire.NewSet(
	wire.Bind(new(TokenManager), new(*TokenManagerImpl)),
	NewTokenValidatorConfig,
	NewTokenManager,
)
