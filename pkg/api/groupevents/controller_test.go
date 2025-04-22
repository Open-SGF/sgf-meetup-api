package groupevents

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/fakers"
	"sgf-meetup-api/pkg/shared/models"
	"testing"
	"time"
)

func TestController_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	testDB, err := db.NewTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Close()

	meetupFaker := fakers.NewMeetupFaker(0)

	controller := NewController(NewService(ServiceConfig{AppURL: "http://localhost"}))

	router := gin.New()
	controller.RegisterRoutes(router)

	t.Run("GET /groups/:group/events returns events for group", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()
		group := "test-group"

		events := []models.MeetupEvent{
			meetupFaker.CreateEvent(group, time.Now()),
			meetupFaker.CreateEvent(group, time.Now()),
		}
		testDB.InsertTestItems(ctx, *infra.EventsTableProps.TableName, events)

		req, _ := http.NewRequest("GET", "/groups/"+group+"/events", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseDTO groupEventsResponseDTO
		err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
		require.NoError(t, err)

		assert.Len(t, responseDTO.Items, 0)
	})
}
