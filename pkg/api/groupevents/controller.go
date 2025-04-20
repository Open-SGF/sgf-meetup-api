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
	r.GET("/groups/:group/events", c.getGroupEvents)
	r.GET("/groups/:group/events/next", c.getNextGroupEvent)
	r.GET("/groups/:group/events/:eventId", c.getGroupEventById)
}

func (c *Controller) getGroupEvents(ctx *gin.Context) {

}

func (c *Controller) getNextGroupEvent(ctx *gin.Context) {

}

func (c *Controller) getGroupEventById(ctx *gin.Context) {

}

var Providers = wire.NewSet(NewController)
