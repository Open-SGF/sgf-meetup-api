//go:generate go tool swag init -g router.go
//go:generate go tool swag fmt

package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"sgf-meetup-api/pkg/api/auth"
	_ "sgf-meetup-api/pkg/api/docs"
	"sgf-meetup-api/pkg/api/groupevents"
)

//	@title		SGF Meetup API
//	@version	1.0

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and the JWT token.

func NewRouter(
	logger *slog.Logger,
	authController *auth.Controller,
	groupEventsController *groupevents.Controller,
	authMiddleware *auth.Middleware,
) *gin.Engine {
	r := gin.Default()

	r.Use(sloggin.New(logger.WithGroup("http")))
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1Group := r.Group("v1")

	authController.RegisterRoutes(v1Group)

	authGroup := v1Group.Group("/")
	authGroup.Use(authMiddleware.Handler)

	groupEventsController.RegisterRoutes(authGroup)

	return r
}
