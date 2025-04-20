package apierrors

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProblemDetails(t *testing.T) {
	t.Run("custom type preserves title", func(t *testing.T) {
		pd := NewProblemDetails(
			http.StatusUnauthorized,
			"https://example.com/auth-error",
			"Invalid credentials",
			"Missing Authorization header",
			"/account/123",
		)

		assert.Equal(t, "https://example.com/auth-error", pd.Type)
		assert.Equal(t, "Invalid credentials", pd.Title)
		assert.Equal(t, http.StatusUnauthorized, pd.Status)
	})

	t.Run("about:blank type uses status text for title", func(t *testing.T) {
		pd := NewProblemDetails(
			http.StatusNotFound,
			"about:blank",
			"Should be overwritten",
			"User not found",
			"",
		)

		assert.Equal(t, http.StatusText(http.StatusNotFound), pd.Title)
	})

	t.Run("empty type defaults to about:blank with status title", func(t *testing.T) {
		pd := NewProblemDetails(
			http.StatusForbidden,
			"",
			"Original title",
			"Access denied",
			"",
		)

		assert.Equal(t, "about:blank", pd.Type)
		assert.Equal(t, http.StatusText(http.StatusForbidden), pd.Title)
	})
}

func TestNewHTTPProblemDetails(t *testing.T) {
	t.Run("creates minimal problem details", func(t *testing.T) {
		pd := NewHTTPProblemDetails(http.StatusBadGateway)

		require.NotNil(t, pd)
		assert.Equal(t, "about:blank", pd.Type)
		assert.Equal(t, http.StatusText(http.StatusBadGateway), pd.Title)
		assert.Equal(t, http.StatusBadGateway, pd.Status)
	})
}

func TestWriteProblemDetails(t *testing.T) {
	t.Run("sets proper headers and body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		pd := &ProblemDetails{
			Type:   "test-error",
			Title:  "Test Error",
			Status: http.StatusTeapot,
		}

		WriteProblemDetails(c, pd)

		assert.Equal(t, http.StatusTeapot, w.Code)
		assert.Equal(t, "application/problem+json", w.Header().Get("Content-Type"))
		assert.JSONEq(t, `{
			"type": "test-error",
			"title": "Test Error",
			"status": 418
		}`, w.Body.String())
	})
}

func TestWriteProblemDetailsFromStatus(t *testing.T) {
	t.Run("generic status-based problem", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		WriteProblemDetailsFromStatus(c, http.StatusRequestTimeout)

		assert.Equal(t, http.StatusRequestTimeout, w.Code)
		assert.JSONEq(t, `{
			"type": "about:blank",
			"title": "Request Timeout",
			"status": 408
		}`, w.Body.String())
	})
}

func TestProblemDetails_GetStatus(t *testing.T) {
	pd := &ProblemDetails{Status: http.StatusNotImplemented}
	assert.Equal(t, http.StatusNotImplemented, pd.GetStatus())
}
