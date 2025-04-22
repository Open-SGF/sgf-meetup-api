package groupevents

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"net/http"
	"net/url"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/api/apierrors"
	"strconv"
	"strings"
	"time"
)

type ControllerConfig struct {
	AppURL url.URL
}

func NewControllerConfig(config *apiconfig.Config) ControllerConfig {
	return ControllerConfig{
		AppURL: config.AppURL,
	}
}

type Controller struct {
	config  ControllerConfig
	service *Service
}

const (
	groupIDKey = "groupId"
	eventIDKey = "eventId"
	cursorKey  = "cursor"
	limitKey   = "limit"
	beforeKey  = "before"
	afterKey   = "after"
)

func NewController(config ControllerConfig, service *Service) *Controller {
	return &Controller{
		config:  config,
		service: service,
	}
}

func (c *Controller) RegisterRoutes(r gin.IRouter) {
	r.GET("/groups/:"+groupIDKey+"/events", c.groupEvents)
	r.GET("/groups/:"+groupIDKey+"/events/next", c.nextGroupEvent)
	r.GET("/groups/:"+groupIDKey+"/events/:"+eventIDKey, c.groupEventByID)
}

// @Summary	Get group events
// @Tags		groupevents
// @Accept		json
// @Produce	json,application/problem+json
// @Param		id		path		string	true	"Group ID"
// @Param		before	query		string	false	"Filter events before this timestamp" Format(date-time)
// @Param		after	query		string	false	"Filter events after this timestamp" Format(date-time)
// @Param		cursor	query		string	false	"Pagination cursor"
// @Param		limit	query		integer	false	"Maximum number of results"
// @Success	200		{object}	groupEventsResponseDTO
// @Failure	400		{object}	apierrors.ProblemDetails	"Invalid input"
// @Failure	401		{object}	apierrors.ProblemDetails	"Unauthorized"
// @Failure	500		{object}	apierrors.ProblemDetails	"Server error"
// @Router		/v1/groups/{groupId}/events [get]
func (c *Controller) groupEvents(ctx *gin.Context) {
	ctx.FullPath()
	groupID := ctx.Param(groupIDKey)

	if groupID == "" {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

	var queryParams groupEventsQueryParams
	if err := ctx.ShouldBindQuery(&queryParams); err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

	events, nextFilters, err := c.service.PaginatedEvents(ctx, groupID, queryParamsToGroupEventArgs(queryParams))

	if err != nil {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, groupEventsResponseDTO{
		Items:       meetupEventsToDTOs(events),
		NextPageURL: c.createNextURL(ctx, groupID, nextFilters),
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
	groupID := ctx.Param(groupIDKey)

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
	groupID := ctx.Param(groupIDKey)
	eventID := ctx.Param(eventIDKey)

	if groupID == "" || eventID == "" {
		apierrors.WriteProblemDetailsFromStatus(ctx, http.StatusBadRequest)
		return
	}

}

func (c *Controller) createNextURL(ctx *gin.Context, groupID string, filters *PaginatedEventsFilters) *string {
	if filters == nil {
		return nil
	}
	path := strings.ReplaceAll(ctx.FullPath(), ":"+groupIDKey, groupID)
	newURL := c.config.AppURL.JoinPath(path)

	query := url.Values{}

	query.Add(cursorKey, filters.Cursor)

	if filters.Limit != nil {
		query.Add(limitKey, strconv.Itoa(*filters.Limit))
	}
	if filters.Before != nil {
		query.Add(beforeKey, filters.Before.Format(time.RFC3339))
	}
	if filters.After != nil {
		query.Add(afterKey, filters.After.Format(time.RFC3339))
	}

	newURL.RawQuery = query.Encode()

	urlString := newURL.String()
	return &urlString
}

var Providers = wire.NewSet(NewService, NewControllerConfig, NewController)
