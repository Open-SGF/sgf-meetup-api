package groupevents

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
	r.GET("/groups/:group/events", c.groupEvents)
	r.GET("/groups/:group/events/next", c.nextGroupEvent)
	r.GET("/groups/:group/events/:eventId", c.groupEventByID)
}

func (c *Controller) groupEvents(ctx *gin.Context) {

}

func (c *Controller) nextGroupEvent(ctx *gin.Context) {

}

func (c *Controller) groupEventByID(ctx *gin.Context) {

}

var Providers = wire.NewSet(NewController)
