package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"net/http"
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
	r.POST("/refresh", c.refresh)
}

// @Summary	Authenticate with credentials
// @Tags		auth
// @Accept		json
// @Produce	json
// @Param		request	body		authRequestDTO	true	"Credentials"
// @Success	200		{object}	authResponseDTO
// @Failure	400		"Invalid input"
// @Failure	401		"Unauthorized"
// @Router		/v1/auth [post]
func (c *Controller) auth(ctx *gin.Context) {
	requestDTO := authRequestDTO{}

	if err := ctx.ShouldBindJSON(&requestDTO); err != nil {
		ctx.String(http.StatusBadRequest, "")
		return
	}

	result, err := c.service.AuthClientCredentials(requestDTO.ClientID, requestDTO.ClientSecret)

	if err != nil {
		ctx.String(http.StatusInternalServerError, "")
		return
	}

	responseDTO := authResponseDTO{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}
	ctx.JSON(http.StatusOK, responseDTO)
}

// @Summary	Refresh token
// @Tags		auth
// @Accept		json
// @Produce	json
// @Param		request	body		refreshTokenRequestDTO	true	"Refresh token"
// @Success	200		{object}	authResponseDTO
// @Failure	400		"Invalid token"
// @Router		/v1/refresh [post]
func (c *Controller) refresh(ctx *gin.Context) {
	requestDTO := refreshTokenRequestDTO{}

	if err := ctx.ShouldBindJSON(&requestDTO); err != nil {
		ctx.String(http.StatusBadRequest, "")
		return
	}

	responseDTO := authResponseDTO{}
	ctx.JSON(http.StatusOK, responseDTO)
}

var ProviderSet = wire.NewSet(NewController, NewService)
