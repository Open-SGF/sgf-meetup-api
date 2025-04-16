//go:generate go tool swag init -g router.go
//go:generate go tool swag fmt

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"sgf-meetup-api/pkg/api/auth"
	_ "sgf-meetup-api/pkg/api/docs"
	"sgf-meetup-api/pkg/api/groupevents"
)

//	@title		SGF Meetup API
//	@version	1.0

func NewRouter(
	authController *auth.Controller,
	groupEventsController *groupevents.Controller,
) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("v1")

	authController.RegisterRoutes(v1)
	groupEventsController.RegisterRoutes(v1)

	return r
}
