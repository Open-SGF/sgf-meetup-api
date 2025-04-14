package models

import "time"

type MeetupEvent struct {
	ID          string        `json:"id" dynamodbav:"id" fake:"{uuid}"`
	GroupId     string        `json:"group.urlname" dynamodbav:"groupId" fake:"{username}"`
	GroupName   string        `json:"group.name" dynamodbav:"groupName" fake:"{username}"`
	Title       string        `json:"title" dynamodbav:"title" fake:"{sentence:3}"`
	EventUrl    string        `json:"eventUrl" dynamodbav:"eventUrl" fake:"{url}"`
	Description string        `json:"description" dynamodbav:"description" fake:"{paragraph:3,5,2,\n}"`
	DateTime    *time.Time    `json:"dateTime" dynamodbav:"dateTime" fake:"{futuredate}"`
	Duration    string        `json:"duration" dynamodbav:"duration" fake:"{randomstring:[2h,1h30m,3h]}"`
	Venue       *MeetupVenue  `json:"venue" dynamodbav:"venue"`
	Host        *MeetupHost   `json:"host" dynamodbav:"host"`
	Images      []MeetupImage `json:"images" dynamodbav:"images" fakesize:"1,3"`
}

type MeetupVenue struct {
	Name       string `json:"name" dynamodbav:"name" fake:"{company}"`
	Address    string `json:"address" dynamodbav:"address" fake:"{street}"`
	City       string `json:"city" dynamodbav:"city" fake:"{city}"`
	State      string `json:"state" dynamodbav:"state" fake:"{state}"`
	PostalCode string `json:"postalCode" dynamodbav:"postalCode" fake:"{zip}"`
}

type MeetupHost struct {
	Name string `json:"name" dynamodbav:"name" fake:"{name}"`
}

type MeetupImage struct {
	BaseUrl string `json:"baseUrl" dynamodbav:"baseUrl" fake:"{url}.{randomstring:[jpg,jpeg,png,svg,webp]}"`
	Preview string `json:"preview" dynamodbav:"preview" fake:"{url}.{randomstring:[jpg,jpeg,png,svg,webp]}"`
}
