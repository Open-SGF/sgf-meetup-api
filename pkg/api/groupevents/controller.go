package groupevents

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"net/http"
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
	r.GET("/groups/:groupId/events", c.groupEvents)
	r.GET("/groups/:groupId/events/next", c.nextGroupEvent)
	r.GET("/groups/:groupId/events/:eventId", c.groupEventByID)
}

// @Summary	Get group events
// @Tags		groupevents
// @Accept		json
// @Produce	json,application/problem+json
// @Param		id		path		string	true	"Group ID"
// @Param		before	query		string	false	"Filter events before this timestamp"
// @Param		after	query		string	false	"Filter events after this timestamp"
// @Param		cursor	query		string	false	"Pagination cursor"
// @Param		limit	query		integer	false	"Maximum number of results"
// @Success	200		{object}	groupEventsResponseDTO
// @Failure	400		{object}	apierrors.ProblemDetails	"Invalid input"
// @Failure	401		{object}	apierrors.ProblemDetails	"Unauthorized"
// @Failure	500		{object}	apierrors.ProblemDetails	"Server error"
// @Router		/v1/groups/{groupId}/events [get]
func (c *Controller) groupEvents(ctx *gin.Context) {
	groupID := ctx.Param("groupId")

	if groupID == "" {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

	var queryParams groupEventsQueryParams
	if err := ctx.ShouldBindQuery(&queryParams); err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

	events, nextURL, err := c.service.GroupEvents(ctx, groupID, queryParamsToGroupEventArgs(queryParams))

	if err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, groupEventsResponseDTO{
		Items:       meetupEventsToDTOs(events),
		NextPageURL: nextURL,
	})
}

// @Summary	Get next group event
// @Tags		groupevents
// @Accept		json
// @Produce	json,application/problem+json
// @Param		groupId	path		string	true	"Group ID"
// @Success	200		{object}	eventDTO
// @Failure	400		{object}	apierrors.ProblemDetails	"Invalid input"
// @Failure	401		{object}	apierrors.ProblemDetails	"Unauthorized"
// @Failure	404		{object}	apierrors.ProblemDetails	"Not found"
// @Failure	500		{object}	apierrors.ProblemDetails	"Server error"
// @Router		/v1/groups/{groupId}/events/next [get]
func (c *Controller) nextGroupEvent(ctx *gin.Context) {
	groupID := ctx.Param("groupId")

	if groupID == "" {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

}

// @Summary	Get group event by ID
// @Tags		groupevents
// @Accept		json
// @Produce	json,application/problem+json
// @Param		groupId	path		string	true	"Group ID"
// @Param		eventId	path		string	true	"Event ID"
// @Success	200		{object}	eventDTO
// @Failure	400		{object}	apierrors.ProblemDetails	"Invalid input"
// @Failure	401		{object}	apierrors.ProblemDetails	"Unauthorized"
// @Failure	404		{object}	apierrors.ProblemDetails	"Not found"
// @Failure	500		{object}	apierrors.ProblemDetails	"Server error"
// @Router		/v1/groups/{groupId}/events/{eventId} [get]
func (c *Controller) groupEventByID(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	eventID := ctx.Param("eventId")

	if groupID == "" || eventID == "" {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

}

var Providers = wire.NewSet(NewServiceConfig, NewService, NewController)
