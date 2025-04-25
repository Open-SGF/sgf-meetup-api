package groupevents

import (
	"sgf-meetup-api/pkg/shared/models"
)

func meetupEventToDTO(meetupEvent *models.MeetupEvent) *eventDTO {
	if meetupEvent == nil {
		return nil
	}

	var date *string
	if meetupEvent.DateTime != nil {
		dateStr := meetupEvent.DateTime.String()
		date = &dateStr
	}

	return &eventDTO{
		ID:          meetupEvent.ID,
		GroupID:     meetupEvent.GroupID,
		GroupName:   meetupEvent.GroupName,
		Title:       meetupEvent.Title,
		EventURL:    meetupEvent.EventURL,
		Description: meetupEvent.Description,
		DateTime:    date,
		Duration:    meetupEvent.Duration,
		Venue:       meetupVenueToDTO(meetupEvent.Venue),
		Host:        meetupHostToDTO(meetupEvent.Host),
		Images:      meetupImagesToDTOs(meetupEvent.Images),
	}
}

func meetupEventsToDTOs(meetupEvents []models.MeetupEvent) []eventDTO {
	dtos := make([]eventDTO, len(meetupEvents))

	for i := range meetupEvents {
		dtos[i] = *meetupEventToDTO(&meetupEvents[i])
	}
	return dtos
}

func meetupVenueToDTO(meetupVenue *models.MeetupVenue) *venueDTO {
	if meetupVenue == nil {
		return nil
	}

	return &venueDTO{
		Name:       meetupVenue.Name,
		Address:    meetupVenue.Address,
		City:       meetupVenue.City,
		State:      meetupVenue.State,
		PostalCode: meetupVenue.PostalCode,
	}
}

func meetupHostToDTO(meetupHost *models.MeetupHost) *hostDTO {
	if meetupHost == nil {
		return nil
	}

	return &hostDTO{
		Name: meetupHost.Name,
	}
}

func meetupImageToDTO(meetupImage *models.MeetupImage) *imageDTO {
	if meetupImage == nil {
		return nil
	}

	return &imageDTO{
		BaseUrl: meetupImage.BaseUrl,
		Preview: meetupImage.Preview,
	}
}

func meetupImagesToDTOs(meetupImages []models.MeetupImage) []imageDTO {
	dtos := make([]imageDTO, len(meetupImages))

	for i := range meetupImages {
		dtos[i] = *meetupImageToDTO(&meetupImages[i])
	}

	return dtos
}

func queryParamsToGroupEventArgs(queryParams groupEventsQueryParams) PaginatedEventsFilters {
	return PaginatedEventsFilters(queryParams)
}
