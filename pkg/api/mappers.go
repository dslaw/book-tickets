package api

import "github.com/dslaw/book-tickets/pkg/entities"

func MapToVenue(data WriteVenueRequest) entities.Venue {
	return entities.Venue{
		Name:        data.Name,
		Description: data.Description,
		Location: entities.VenueLocation{
			Address:     data.Location.Address,
			City:        data.Location.City,
			Subdivision: data.Location.Subdivision,
			CountryCode: data.Location.CountryCode,
		},
	}
}

func MapToVenueResponse(venue entities.Venue) GetVenueResponse {
	response := GetVenueResponse{
		ID:          venue.ID,
		Name:        venue.Name,
		Description: venue.Description,
	}
	response.Location.Address = venue.Location.Address
	response.Location.City = venue.Location.City
	response.Location.Subdivision = venue.Location.Subdivision
	response.Location.CountryCode = venue.Location.CountryCode
	return response
}

func MapToEvent(data WriteEventRequest) entities.Event {
	event := entities.Event{
		Name:        data.Name,
		Description: data.Description,
		StartsAt:    data.StartsAt,
		EndsAt:      data.EndsAt,
		Venue:       entities.EventVenue{ID: data.VenueID},
		Performers:  make([]entities.Performer, len(data.Performers)),
	}

	for idx, performer := range data.Performers {
		event.Performers[idx] = entities.Performer{Name: performer.Name}
	}
	return event
}

func MapToEventResponse(event entities.Event) GetEventResponse {
	response := GetEventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		StartsAt:    event.StartsAt,
		EndsAt:      event.EndsAt,
		Venue: EventVenueResponse{
			ID:   event.Venue.ID,
			Name: event.Venue.Name,
		},
		Performers: make([]EventPerformerResponse, len(event.Performers)),
	}

	for idx, performer := range event.Performers {
		response.Performers[idx] = EventPerformerResponse{ID: performer.ID, Name: performer.Name}
	}
	return response
}
