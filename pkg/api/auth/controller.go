package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) RegisterRoutes(r gin.IRouter) {
	r.POST("/auth", c.auth)
	r.POST("/refresh", c.refresh)
}

//	@Summary	Authenticate with credentials
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		authRequestDTO	true	"Credentials"
//	@Success	200		{object}	authResponseDTO
//	@Failure	400		"Invalid input"
//	@Failure	401		"Unauthorized"
//	@Router		/v1/auth [post]
func (c *Controller) auth(ctx *gin.Context) {

}

//	@Summary	Refresh token
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		refreshTokenRequestDTO	true	"Refresh token"
//	@Success	200		{object}	authResponseDTO
//	@Failure	400		"Invalid token"
//	@Router		/v1/refresh [post]
func (c *Controller) refresh(context *gin.Context) {

}

var ProviderSet = wire.NewSet(NewController)
