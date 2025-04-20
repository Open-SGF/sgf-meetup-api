package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"sgf-meetup-api/pkg/api/apierrors"
	"strings"
)

const ClientIDKey = "clientId"

type Middleware struct {
	tokenValidator TokenManager
}

func NewMiddleware(tokenValidator TokenManager) *Middleware {
	return &Middleware{
		tokenValidator: tokenValidator,
	}
}

func (m *Middleware) Handler(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusUnauthorized)
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusUnauthorized)
		return
	}

	token, err := m.tokenValidator.Validate(tokenParts[1])

	if errors.Is(err, ErrInvalidCredentials) {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusUnauthorized)
		return
	}

	if err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusInternalServerError)
		return
	}

	ctx.Set(ClientIDKey, token.ClientID)
	ctx.Next()
}
