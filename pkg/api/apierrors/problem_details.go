package apierrors

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ProblemDetailer interface {
	GetStatus() int
}

type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func NewProblemDetails(statusCode int, problemType, title, detail, instance string) *ProblemDetails {
	if problemType == "" {
		problemType = "about:blank"
	}

	if problemType == "about:blank" {
		title = http.StatusText(statusCode)
	}

	return &ProblemDetails{
		Type:     problemType,
		Title:    title,
		Status:   statusCode,
		Detail:   detail,
		Instance: instance,
	}
}

func NewHTTPProblemDetails(statusCode int) *ProblemDetails {
	return NewProblemDetails(statusCode, "", "", "", "")
}

func (pd *ProblemDetails) GetStatus() int {
	return pd.Status
}

func WriteProblemDetails(ctx *gin.Context, pd ProblemDetailer) {
	ctx.Header("Content-Type", "application/problem+json")
	ctx.JSON(pd.GetStatus(), pd)
}

func WriteProblemDetailsFromStatus(ctx *gin.Context, status int) {
	pd := NewHTTPProblemDetails(status)
	WriteProblemDetails(ctx, pd)
}
