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
