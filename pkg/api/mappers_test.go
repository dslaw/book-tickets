package api_test

import (
	"testing"

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
