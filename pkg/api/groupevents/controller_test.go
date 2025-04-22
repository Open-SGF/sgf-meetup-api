package groupevents

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
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

	t.Run("GET /groups/:groupId/events returns future events for group", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()
		group := "test-group"

		events := []models.MeetupEvent{
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*1)),
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*2)),
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*-1)),
			meetupFaker.CreateEvent("other-group", timeSource.Now().Add(time.Hour*3)),
		}
		testDB.InsertTestItems(ctx, *infra.EventsTableProps.TableName, events)

		req, _ := http.NewRequest("GET", "/groups/"+group+"/events", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		responseDTO := getDTOWhenStatus[groupEventsResponseDTO](t, w, http.StatusOK)

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

		w := makeRequest(router, "GET", "/groups/"+group+"/events?limit=10", nil)
		responseDTO := getDTOWhenStatus[groupEventsResponseDTO](t, w, http.StatusOK)

		assert.Equal(t, 10, len(responseDTO.Items))
		require.NotNil(t, responseDTO.NextPageURL)

		w = makeRequest(router, "GET", *responseDTO.NextPageURL, nil)
		nextResponseDTO := getDTOWhenStatus[groupEventsResponseDTO](t, w, http.StatusOK)

		assert.Equal(t, 5, len(nextResponseDTO.Items))
		assert.Nil(t, nextResponseDTO.NextPageURL)
	})

	t.Run("GET /groups/:groupId/events handles", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()
		group := "test-group"

		events := []models.MeetupEvent{
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*-2)),
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*-1)),
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*1)),
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*2)),
			meetupFaker.CreateEvent(group, timeSource.Now().Add(time.Hour*3)),
		}

		testDB.InsertTestItems(ctx, *infra.EventsTableProps.TableName, events)

		t.Run("before filter", func(t *testing.T) {
			before := timeSource.Now().Add(time.Hour * 2).UTC().Format(time.RFC3339)
			w := makeRequest(router, "GET", "/groups/"+group+"/events?before="+url.QueryEscape(before), nil)
			responseDTO := getDTOWhenStatus[groupEventsResponseDTO](t, w, http.StatusOK)

			assert.Equal(t, 4, len(responseDTO.Items))
		})

		t.Run("after filter", func(t *testing.T) {
			after := timeSource.Now().Add(time.Hour * 2).Add(time.Second * -1).UTC().Format(time.RFC3339)
			w := makeRequest(router, "GET", "/groups/"+group+"/events?after="+url.QueryEscape(after), nil)
			responseDTO := getDTOWhenStatus[groupEventsResponseDTO](t, w, http.StatusOK)

			assert.Equal(t, 2, len(responseDTO.Items))
		})

		t.Run("before and after filter", func(t *testing.T) {
			before := timeSource.Now().Add(time.Hour * 2).UTC().Format(time.RFC3339)
			after := timeSource.Now().Add(time.Hour * -1).Add(time.Second * -1).UTC().Format(time.RFC3339)
			w := makeRequest(router, "GET", "/groups/"+group+"/events?after="+url.QueryEscape(after)+"&before="+url.QueryEscape(before), nil)
			responseDTO := getDTOWhenStatus[groupEventsResponseDTO](t, w, http.StatusOK)

			assert.Equal(t, 3, len(responseDTO.Items))
		})
	})
}

func makeRequest(router *gin.Engine, method, url string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func getDTOWhenStatus[T any](t *testing.T, w *httptest.ResponseRecorder, status int) T {
	require.Equal(t, status, w.Code)
	var dto T
	err := json.Unmarshal(w.Body.Bytes(), &dto)
	require.NoError(t, err)
	return dto
}
