package api

import "github.com/gin-gonic/gin"

func Router() *gin.Engine {
	r := gin.Default()

	r.GET("/events", getEvents)

	r.GET("/events/next", getNextEvent)

	return r
}
