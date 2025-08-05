package groupevents

import (
	"time"
)

type groupEventsQueryParams struct {
	Before *time.Time `form:"before"`
	After  *time.Time `form:"after"`
	Cursor string     `form:"cursor"`
	Limit  *int       `form:"limit"`
}

type groupEventsResponseDTO struct {
	Items       []eventDTO `json:"items"`
	NextPageURL *string    `json:"nextPageUrl"`
}

type eventDTO struct {
	ID          string     `json:"id"`
	Group       groupDTO   `json:"group"`
	Title       string     `json:"title"`
	EventURL    string     `json:"eventUrl"`
	Description string     `json:"description"`
	DateTime    *time.Time `json:"dateTime"`
	Duration    string     `json:"duration"`
	Venue       *venueDTO  `json:"venue"`
	Host        *hostDTO   `json:"host"`
	Images      []imageDTO `json:"images"`
}

type groupDTO struct {
	URLName string `json:"urlname"`
	Name    string `json:"name"`
}

type venueDTO struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
}

type hostDTO struct {
	Name string `json:"name"`
}

type imageDTO struct {
	BaseURL string `json:"baseUrl"`
	Preview string `json:"preview"`
}
