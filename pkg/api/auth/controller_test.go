package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/models"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestController_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	testDB, err := db.NewTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Close()
	timeSource := clock.NewMockTimeSource(time.Now())
	apiUserRepo := NewDynamoDBAPIUserRepository(DynamoDBAPIUserRepositoryConfig{
		APIUserTable: *infra.ApiUsersTableProps.TableName,
	}, testDB.Client)
	tokenSecret := []byte("some-secret-value")
	tokenValidator := NewTokenManager(TokenManagerConfig{
		JWTIssuer: "meetup-api.opensgf.org",
		JWTSecret: tokenSecret,
	}, timeSource)
	service := NewService(ServiceConfig{
		AccessTokenExpiration:  time.Minute * 15,
		RefreshTokenExpiration: time.Hour * 24 * 30,
	}, timeSource, apiUserRepo, tokenValidator)
	controller := NewController(service)

	router := gin.New()
	controller.RegisterRoutes(router)

	t.Run("POST /auth creates token", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		requestDTO := authRequestDTO{
			ClientID:     "someClientId",
			ClientSecret: "someClientSecret",
		}

		addAPIUser(t, ctx, testDB.Client, requestDTO.ClientID, requestDTO.ClientSecret)

		jsonValue, _ := json.Marshal(requestDTO)
		req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseDTO authResponseDTO
		err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
		require.NoError(t, err)

		assert.NotEmpty(t, responseDTO.AccessToken)
		assert.True(t, responseDTO.AccessTokenExpiresAt.After(time.Now()))
		assert.NotEmpty(t, responseDTO.RefreshToken)
		assert.True(t, responseDTO.RefreshTokenExpiresAt.After(time.Now()))

		accessToken, accessTokenClaims, err := parseJWTToken(responseDTO.AccessToken, tokenSecret)
		require.NoError(t, err)
		assert.True(t, accessToken.Valid)
		assert.Equal(t, accessTokenClaims.Subject, requestDTO.ClientID)

		refreshToken, refreshTokenClaims, err := parseJWTToken(
			responseDTO.RefreshToken,
			tokenSecret,
		)
		require.NoError(t, err)
		assert.True(t, refreshToken.Valid)
		assert.Equal(t, refreshTokenClaims.Subject, requestDTO.ClientID)
	})

	t.Run("POST /auth handles invalid credentials", func(t *testing.T) {
		requestDTO := authRequestDTO{
			ClientID:     "invalid",
			ClientSecret: "invalid",
		}
		jsonValue, _ := json.Marshal(requestDTO)
		req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("POST /auth handles invalid client secret for valid client id", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		requestDTO := authRequestDTO{
			ClientID:     "someClientId",
			ClientSecret: "invalidClientSecret",
		}

		addAPIUser(t, ctx, testDB.Client, requestDTO.ClientID, "realClientSecret")

		jsonValue, _ := json.Marshal(requestDTO)
		req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("POST /auth handles invalid json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer([]byte("invalid json")))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /auth/refresh refreshes token", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		clientID := "someClientId"
		clientSecret := "someClientSecret"

		addAPIUser(t, ctx, testDB.Client, clientID, clientSecret)

		result, err := service.AuthClientCredentials(ctx, clientID, clientSecret)
		require.NoError(t, err)

		timeSource.SetTime(time.Now().Add(time.Hour * 24 * 31))
		defer timeSource.Reset()

		requestDTO := refreshTokenRequestDTO{
			RefreshToken: result.RefreshToken,
		}

		jsonValue, _ := json.Marshal(requestDTO)
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("POST /auth/refresh handles expired token", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		clientID := "someClientId"
		clientSecret := "someClientSecret"

		addAPIUser(t, ctx, testDB.Client, clientID, clientSecret)

		result, err := service.AuthClientCredentials(ctx, clientID, clientSecret)
		require.NoError(t, err)

		timeSource.SetTime(time.Now().Add(time.Hour * 24))
		defer timeSource.Reset()

		requestDTO := refreshTokenRequestDTO{
			RefreshToken: result.RefreshToken,
		}

		jsonValue, _ := json.Marshal(requestDTO)
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseDTO authResponseDTO
		err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
		require.NoError(t, err)

		assert.NotEqual(t, result.AccessToken, responseDTO.AccessToken)
		assert.NotEqual(t, result.RefreshToken, responseDTO.RefreshToken)

		assert.NotEmpty(t, responseDTO.AccessToken)
		assert.True(t, responseDTO.AccessTokenExpiresAt.After(time.Now()))
		assert.NotEmpty(t, responseDTO.RefreshToken)
		assert.True(t, responseDTO.RefreshTokenExpiresAt.After(time.Now()))

		accessToken, accessTokenClaims, err := parseJWTToken(responseDTO.AccessToken, tokenSecret)
		require.NoError(t, err)
		assert.True(t, accessToken.Valid)
		assert.Equal(t, accessTokenClaims.Subject, clientID)

		refreshToken, refreshTokenClaims, err := parseJWTToken(
			responseDTO.RefreshToken,
			tokenSecret,
		)
		require.NoError(t, err)
		assert.True(t, refreshToken.Valid)
		assert.Equal(t, refreshTokenClaims.Subject, clientID)
	})

	t.Run("POST /auth/refresh handles invalid token", func(t *testing.T) {
		requestDTO := refreshTokenRequestDTO{
			RefreshToken: "invalid",
		}
		jsonValue, _ := json.Marshal(requestDTO)
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("POST /auth/refresh handles invalid json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer([]byte("invalid json")))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func addAPIUser(t *testing.T, ctx context.Context, client *db.Client, id, secret string) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	require.NoError(t, err)

	apiUser := models.APIUser{
		ClientID:           id,
		HashedClientSecret: bytes,
	}

	av, err := attributevalue.MarshalMap(
		apiUser,
	) // Use Marshal to convert struct to AttributeValues
	require.NoError(t, err)

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: infra.ApiUsersTableProps.TableName,
		Item:      av,
	})

	require.NoError(t, err)
}

func parseJWTToken(tokenString string, secret []byte) (*jwt.Token, *jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return secret, nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, nil, errors.New("invalid token claims")
	}

	return token, claims, nil
}
