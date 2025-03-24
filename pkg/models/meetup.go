package models

import "time"

type MeetupEvent struct {
	Id          string        `json:"id"`
	Title       string        `json:"title"`
	EventUrl    string        `json:"eventUrl"`
	Description string        `json:"description"`
	DateTime    *time.Time    `json:"dateTime"`
	Duration    string        `json:"duration"`
	Venue       *MeetupVenue  `json:"venue"`
	Group       *MeetupGroup  `json:"group"`
	Host        *MeetupHost   `json:"host"`
	Images      []MeetupImage `json:"images"`
	DeletedAt   *time.Time    `json:"deletedAt"`
}

type MeetupVenue struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
}

type MeetupGroup struct {
	Name    string `json:"name"`
	UrlName string `json:"urlname"`
}

type MeetupHost struct {
	Name string `json:"name"`
}

type MeetupImage struct {
	BaseUrl string `json:"baseUrl"`
	Preview string `json:"preview"`
}
