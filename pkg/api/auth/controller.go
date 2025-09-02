package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/api/apierrors"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

func (c *Controller) RegisterRoutes(r gin.IRouter) {
	r.POST("/auth", c.auth)
	r.POST("/auth/refresh", c.refresh)
}

// @Summary	Authenticate with credentials
// @Tags		auth
// @Accept		json
// @Produce	json,application/problem+json
// @Param		request	body		authRequestDTO	true	"Credentials"
// @Success	200		{object}	authResponseDTO
// @Failure	400		{object}	apierrors.ProblemDetails	"Invalid input"
// @Failure	401		{object}	apierrors.ProblemDetails	"Unauthorized"
// @Failure	500		{object}	apierrors.ProblemDetails	"Server error"
// @Router		/v1/auth [post]
func (c *Controller) auth(ctx *gin.Context) {
	requestDTO := authRequestDTO{}

	if err := ctx.ShouldBindJSON(&requestDTO); err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

	result, err := c.service.AuthClientCredentials(
		ctx,
		requestDTO.ClientID,
		requestDTO.ClientSecret,
	)

	if errors.Is(err, ErrInvalidCredentials) {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusUnauthorized)
		return
	}

	if err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, &authResponseDTO{
		AccessToken:           result.AccessToken,
		AccessTokenExpiresAt:  result.AccessTokenExpiresAt,
		RefreshToken:          result.RefreshToken,
		RefreshTokenExpiresAt: result.RefreshTokenExpiresAt,
	})
}

// @Summary	Refresh token
// @Tags		auth
// @Accept		json
// @Produce	json,application/problem+json
// @Param		request	body		refreshTokenRequestDTO	true	"Refresh token"
// @Success	200		{object}	authResponseDTO
// @Failure	400		{object}	apierrors.ProblemDetails	"Invalid input"
// @Failure	401		{object}	apierrors.ProblemDetails	"Unauthorized"
// @Failure	500		{object}	apierrors.ProblemDetails	"Server error"
// @Router		/v1/auth/refresh [post]
func (c *Controller) refresh(ctx *gin.Context) {
	requestDTO := refreshTokenRequestDTO{}

	if err := ctx.ShouldBindJSON(&requestDTO); err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

	result, err := c.service.RefreshCredentials(ctx, requestDTO.RefreshToken)

	if errors.Is(err, ErrInvalidCredentials) {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusUnauthorized)
		return
	}

	if err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, authResponseDTO{
		AccessToken:           result.AccessToken,
		AccessTokenExpiresAt:  result.AccessTokenExpiresAt,
		RefreshToken:          result.RefreshToken,
		RefreshTokenExpiresAt: result.RefreshTokenExpiresAt,
	})
}

var Providers = wire.NewSet(
	APIUserRepositoryProviders,
	TokenValidatorProviders,
	NewServiceConfig,
	NewService,
	NewController,
	NewMiddleware,
)
