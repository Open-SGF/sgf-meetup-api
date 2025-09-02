package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/shared/clock"
)

func TestMiddleware_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	timeSource := clock.NewMockTimeSource(time.Now())
	tokenManager := NewTokenManager(TokenManagerConfig{
		JWTIssuer: "issuer",
		JWTSecret: []byte("secret"),
	}, timeSource)

	middleware := NewMiddleware(tokenManager)

	t.Run("should return 401 when no authorization header is present", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		middleware.Handler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("should return 401 when authorization header is invalid format", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "invalid_format")
		middleware.Handler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("should return 401 when token is invalid", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid_token")
		middleware.Handler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("should set client ID in context when token is valid", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		token, err := tokenManager.CreateSignedToken("test_client", time.Now().Add(time.Hour))
		require.NoError(t, err)

		c.Request.Header.Set("Authorization", "Bearer "+token)
		middleware.Handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		clientID, _ := c.Get(ClientIDKey)
		assert.Equal(t, "test_client", clientID)
		assert.False(t, c.IsAborted())
	})
}
