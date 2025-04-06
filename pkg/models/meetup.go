package models

import "time"

type MeetupEvent struct {
	Id          string        `json:"id" fake:"{uuid}"`
	Title       string        `json:"title" fake:"{sentence:3}"`
	EventUrl    string        `json:"eventUrl" fake:"{url}"`
	Description string        `json:"description" fake:"{paragraph:3,5,2,\n}"`
	DateTime    *time.Time    `json:"dateTime" fake:"{futuredate}"`
	Duration    string        `json:"duration" fake:"{randomstring:[2h,1h30m,3h]}"`
	Venue       *MeetupVenue  `json:"venue"`
	Group       *MeetupGroup  `json:"group"`
	Host        *MeetupHost   `json:"host"`
	Images      []MeetupImage `json:"images" fakesize:"1,3"`
	DeletedAt   *time.Time    `json:"deletedAt" fake:"skip"`
}

type MeetupVenue struct {
	Name       string `json:"name" fake:"{company}"`
	Address    string `json:"address" fake:"{street}"`
	City       string `json:"city" fake:"{city}"`
	State      string `json:"state" fake:"{state}"`
	PostalCode string `json:"postalCode" fake:"{zip}"`
}

type MeetupGroup struct {
	Name    string `json:"name" fake:"{appname}"`
	UrlName string `json:"urlname" fake:"{username}"`
}

type MeetupHost struct {
	Name string `json:"name" fake:"{name}"`
}

type MeetupImage struct {
	BaseUrl string `json:"baseUrl" fake:"{url}.{randomstring:[jpg,jpeg,png,svg,webp]}"`
	Preview string `json:"preview" fake:"{url}.{randomstring:[jpg,jpeg,png,svg,webp]}"`
}
