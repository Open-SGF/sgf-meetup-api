package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeetupEvent_UnmarshalJSON(t *testing.T) {
	jsonStr := `
{
  "id": "298237443",
  "title": "Code & Demo Night",
  "eventUrl": "https://www.meetup.com/open-sgf/events/298237443",
  "description": "some description",
  "dateTime": "2024-01-16T18:30-06:00",
  "duration": "PT2H",
  "venue": {
    "name": "efactory",
    "address": "405 N Jefferson Ave",
    "city": "Springfield",
    "state": "MO",
    "postalCode": "65806"
  },
  "group": {
    "name": "Open SGF",
    "urlname": "open-sgf"
  },
  "host": {
    "name": "Levi Zitting"
  },
  "images": [
    {
      "baseUrl": "https://secure-content.meetupstatic.com/images/classic-events/",
      "preview": null
    }
  ]
}`

	var event MeetupEvent
	err := json.Unmarshal([]byte(jsonStr), &event)
	require.NoError(t, err)

	assert.Equal(t, "298237443", event.ID)
	assert.Equal(t, "Code & Demo Night", event.Title)
	assert.Equal(t, "https://www.meetup.com/open-sgf/events/298237443", event.EventURL)
	assert.Equal(t, "some description", event.Description)
	assert.Equal(t, "2024-01-16T18:30-06:00", event.DateTime.String())
	assert.Equal(t, "PT2H", event.Duration)
	assert.Equal(t, "efactory", event.Venue.Name)
	assert.Equal(t, "405 N Jefferson Ave", event.Venue.Address)
	assert.Equal(t, "Springfield", event.Venue.City)
	assert.Equal(t, "MO", event.Venue.State)
	assert.Equal(t, "65806", event.Venue.PostalCode)
	assert.Equal(t, "Open SGF", event.GroupName)
	assert.Equal(t, "open-sgf", event.GroupID)
	assert.Equal(t, "Levi Zitting", event.Host.Name)
	assert.Len(t, event.Images, 1)
	assert.Equal(
		t,
		"https://secure-content.meetupstatic.com/images/classic-events/",
		event.Images[0].BaseUrl,
	)
}
