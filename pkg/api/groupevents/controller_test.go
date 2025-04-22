package groupevents

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/fakers"
	"sgf-meetup-api/pkg/shared/models"
	"testing"
	"time"
)

func TestNewControllerConfig(t *testing.T) {
	u, err := url.ParseRequestURI("/")
	require.NoError(t, err)

	cfg := &apiconfig.Config{
		AppURL: *u,
	}

	controllerConfig := NewControllerConfig(cfg)

	assert.Equal(t, cfg.AppURL, controllerConfig.AppURL)
}

func TestController_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	testDB, err := db.NewTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Close()

	meetupFaker := fakers.NewMeetupFaker(0)

	u, err := url.ParseRequestURI("/")
	require.NoError(t, err)
	timeSource := clock.NewMockTimeSource(time.Now().UTC())
	groupEventRepo := NewDynamoDBGroupEventRepository(DynamoDBGroupEventRepositoryConfig{
		EventsTableName:    *infra.EventsTableProps.TableName,
		GroupDateIndexName: *infra.GroupIdDateTimeIndex.IndexName,
	}, timeSource, testDB.Client)
	controller := NewController(ControllerConfig{AppURL: *u}, groupEventRepo)

	router := gin.New()
	controller.RegisterRoutes(router)

	t.Run("GET /groups/:groupId/events returns events for group", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()
		group := "test-group"

		events := []models.MeetupEvent{
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*1)),
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*2)),
			meetupFaker.CreateEvent("other-group", timeSource.Now().Add(time.Hour*3)),
		}
		testDB.InsertTestItems(ctx, *infra.EventsTableProps.TableName, events)

		req, _ := http.NewRequest("GET", "/groups/"+group+"/events", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseDTO groupEventsResponseDTO
		err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
		require.NoError(t, err)

		assert.Len(t, responseDTO.Items, 2)
		assert.Nil(t, responseDTO.NextPageURL)
		assert.Equal(t, events[0].ID, responseDTO.Items[0].ID)
		assert.Equal(t, events[1].ID, responseDTO.Items[1].ID)
	})

	t.Run("GET /groups/:groupId/events handles pagination", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()
		group := "test-group"

		events := make([]models.MeetupEvent, 15)

		for i := range events {
			events[i] = meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*time.Duration(i+1)))
		}

		testDB.InsertTestItems(ctx, *infra.EventsTableProps.TableName, events)

		req, _ := http.NewRequest("GET", "/groups/"+group+"/events?limit=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseDTO groupEventsResponseDTO
		err = json.Unmarshal(w.Body.Bytes(), &responseDTO)
		require.NoError(t, err)

		assert.Equal(t, 10, len(responseDTO.Items))
		require.NotNil(t, responseDTO.NextPageURL)
		fmt.Println(*responseDTO.NextPageURL)

		req, _ = http.NewRequest("GET", *responseDTO.NextPageURL, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var nextResponseDTO groupEventsResponseDTO
		err = json.Unmarshal(w.Body.Bytes(), &nextResponseDTO)
		require.NoError(t, err)

		assert.Equal(t, 5, len(nextResponseDTO.Items))
		assert.Nil(t, nextResponseDTO.NextPageURL)
	})
}
