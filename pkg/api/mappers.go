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

func MapToTickets(data WriteTicketReleaseRequest, eventID int32) []entities.Ticket {
	totalTickets := 0
	for _, batch := range data.TicketReleases {
		totalTickets += int(batch.Number)
	}

	tickets := make([]entities.Ticket, totalTickets)
	idx := 0
	for _, batch := range data.TicketReleases {
		for range batch.Number {
			tickets[idx] = entities.Ticket{
				EventID: eventID,
				Price:   batch.Price,
				Seat:    batch.Seat,
			}
			idx++
		}
	}

	return tickets
}

func MapToAvailableTicketsAggregateResponse(ticketAggregates []entities.AvailableTicketAggregate) GetAvailableTicketsAggregateResponse {
	aggregates := make([]GetAvailableTicketsAggregate, len(ticketAggregates))
	for idx, ticketAggregate := range ticketAggregates {
		aggregates[idx] = GetAvailableTicketsAggregate{
			Seat:      ticketAggregate.Seat,
			Price:     ticketAggregate.Price,
			TicketIDs: ticketAggregate.IDs,
		}
	}
	return GetAvailableTicketsAggregateResponse{Available: aggregates}
}
