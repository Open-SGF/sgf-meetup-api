package auth

import (
	"errors"
	"net/http"
	"strings"

	"sgf-meetup-api/pkg/api/apierrors"

	"github.com/gin-gonic/gin"
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
		ctx.Abort()
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	token, err := m.tokenValidator.Validate(tokenParts[1])

	if errors.Is(err, ErrInvalidCredentials) {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	if err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusInternalServerError)
		ctx.Abort()
		return
	}

	ctx.Set(ClientIDKey, token.ClientID)
	ctx.Next()
}
