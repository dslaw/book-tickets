package api_test

import (
	"testing"
	"time"

	"github.com/dslaw/book-tickets/pkg/api"
	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestMapToVenue(t *testing.T) {
	requestData := api.WriteVenueRequest{
		Name:        "Test Venue",
		Description: "",
	}
	requestData.Location.Address = "11 Front Street"
	requestData.Location.City = "San Francisco"
	requestData.Location.Subdivision = "CA"
	requestData.Location.CountryCode = "USA"

	expected := entities.Venue{
		Name:        "Test Venue",
		Description: "",
		Location: entities.VenueLocation{
			Address:     "11 Front Street",
			City:        "San Francisco",
			Subdivision: "CA",
			CountryCode: "USA",
		},
	}
	actual := api.MapToVenue(requestData)
	assert.EqualValues(t, expected, actual)
}

func TestMapToVenueResponse(t *testing.T) {
	venue := entities.Venue{
		ID:          1,
		Name:        "Test Venue",
		Description: "",
		Location: entities.VenueLocation{
			Address:     "11 Front Street",
			City:        "San Francisco",
			Subdivision: "CA",
			CountryCode: "USA",
		},
	}
	expected := api.GetVenueResponse{
		ID:          1,
		Name:        "Test Venue",
		Description: "",
	}
	expected.Location.Address = "11 Front Street"
	expected.Location.City = "San Francisco"
	expected.Location.Subdivision = "CA"
	expected.Location.CountryCode = "USA"

	actual := api.MapToVenueResponse(venue)
	assert.EqualValues(t, expected, actual)
}

func TestMapToEvent(t *testing.T) {
	startsAt, _ := time.Parse(time.DateOnly, "2020-01-01")
	endsAt, _ := time.Parse(time.DateOnly, "2020-01-02")

	requestData := api.WriteEventRequest{
		VenueID:     1,
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Performers: []api.WritePerformerRequest{
			{Name: "Performer 1"},
			{Name: "Performer 2"},
		},
	}

	expected := entities.Event{
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Venue:       entities.EventVenue{ID: 1},
		Performers: []entities.Performer{
			{Name: "Performer 1"},
			{Name: "Performer 2"},
		},
	}

	actual := api.MapToEvent(requestData)
	assert.EqualValues(t, expected, actual)
}

func TestMapToEventResponse(t *testing.T) {
	startsAt, _ := time.Parse(time.DateOnly, "2020-01-01")
	endsAt, _ := time.Parse(time.DateOnly, "2020-01-01")

	event := entities.Event{
		ID:          1,
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Venue: entities.EventVenue{
			ID:   1,
			Name: "Test Venue",
		},
		Performers: []entities.Performer{
			{ID: 1, Name: "Test Performer 1"},
			{ID: 2, Name: "Test Performer 2"},
		},
	}
	expected := api.GetEventResponse{
		ID:          1,
		Name:        "Test Event",
		Description: "",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Venue: api.EventVenueResponse{
			ID:   1,
			Name: "Test Venue",
		},
		Performers: []api.EventPerformerResponse{
			{ID: 1, Name: "Test Performer 1"},
			{ID: 2, Name: "Test Performer 2"},
		},
	}

	actual := api.MapToEventResponse(event)
	assert.EqualValues(t, expected, actual)
}

func TestMapToTickets(t *testing.T) {
	eventID := int32(1)
	requestData := api.WriteTicketReleaseRequest{
		TicketReleases: []api.WriteTicketRelease{
			{Number: 2, Seat: "GA", Price: 10},
			{Number: 3, Seat: "Balcony", Price: 20},
		},
	}

	expected := []entities.Ticket{
		{EventID: eventID, Price: 10, Seat: "GA"},
		{EventID: eventID, Price: 10, Seat: "GA"},
		{EventID: eventID, Price: 20, Seat: "Balcony"},
		{EventID: eventID, Price: 20, Seat: "Balcony"},
		{EventID: eventID, Price: 20, Seat: "Balcony"},
	}

	actual := api.MapToTickets(requestData, eventID)
	assert.EqualValues(t, expected, actual)
}

func TestMapToAvailableTicketsAggregateResponse(t *testing.T) {
	ticketAggregates := []entities.AvailableTicketAggregate{
		{Seat: "GA", Price: 10, IDs: []int32{1, 2, 3}},
		{Seat: "Balcony", Price: 20, IDs: []int32{4, 5}},
	}
	expected := api.GetAvailableTicketsAggregateResponse{
		Available: []api.GetAvailableTicketsAggregate{
			{Seat: "GA", Price: 10, TicketIDs: []int32{1, 2, 3}},
			{Seat: "Balcony", Price: 20, TicketIDs: []int32{4, 5}},
		},
	}

	actual := api.MapToAvailableTicketsAggregateResponse(ticketAggregates)
	assert.EqualValues(t, expected, actual)
}
